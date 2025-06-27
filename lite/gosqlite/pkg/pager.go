package pkg

import (
	"fmt"
	"io"
	"sync"

	"gosqlite/pkg/metrics"
)

// Pager manages reading/writing pages from the database file and caching them.
type Pager struct {
	vfs        VFS
	file       File
	pageSize   uint16
	dbSize     uint32 // Current size of the database in pages
	cache      *ARCCache // ARC cache for pages
	dirtyPages map[pkg.PageID]struct{} // Set of page IDs that are dirty
	mu         sync.Mutex // Mutex to protect concurrent access to pager state
	cacheHits  *metrics.Metric
	cacheMisses *metrics.Metric
	// ... other fields for journal/WAL management (in later phases)
}

// NewPager initializes a new Pager.
func NewPager(vfs VFS, file File, pageSize uint16, cacheSize int) (*Pager, error) {
	cacheHits, err := metrics.RegisterCounter("pager_cache_hits")
	if err != nil {
		return nil, fmt.Errorf("failed to register pager_cache_hits metric: %w", err)
	}
	cacheMisses, err := metrics.RegisterCounter("pager_cache_misses")
	if err != nil {
		return nil, fmt.Errorf("failed to register pager_cache_misses metric: %w", err)
	}

	p := &Pager{
		vfs:        vfs,
		file:       file,
		pageSize:   pageSize,
		cache:      NewARCCache(cacheSize), // Initialize ARC cache
		dirtyPages: make(map[pkg.PageID]struct{}),
		cacheHits:  cacheHits,
		cacheMisses: cacheMisses,
	}

	// Determine initial dbSize from file size
	fileSize, err := file.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}
	p.dbSize = uint32(fileSize / int64(pageSize))

	return p, nil
}

// GetPage retrieves a page from the pager. It first checks the cache, then reads from disk.
func (p *Pager) GetPage(id pkg.PageID) (pkg.Page, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if id == 0 { // Page IDs are 1-indexed
		return nil, fmt.Errorf("page ID cannot be 0")
	}

	// Check cache first
	if page, ok := p.cache.Get(id); ok {
		p.cacheHits.Inc()
		return page, nil
	}

	p.cacheMisses.Inc()
	// If not in cache, read from disk
	offset := int64(id-1) * int64(p.pageSize)
	// Zero-Allocation Read: For extreme performance, a pre-allocated buffer pool
	// could be used here to minimize allocations during reads in hot paths.
	page := make(pkg.Page, p.pageSize)
	n, err := p.file.ReadAt(page, offset)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read page %d from disk: %w", id, err)
	}
	if n != int(p.pageSize) && err != io.EOF {
		return nil, fmt.Errorf("short read for page %d: expected %d bytes, got %d", id, p.pageSize, n)
	}

	// Add to cache
	p.cache.Put(id, page)

	return page, nil
}

// WritePage writes a page to the pager. It updates the cache and marks the page as dirty.
// Actual disk writes are deferred to transaction commit.
func (p *Pager) WritePage(id pkg.PageID, data pkg.Page) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if id == 0 { // Page IDs are 1-indexed
		return fmt.Errorf("page ID cannot be 0")
	}
	if uint16(len(data)) != p.pageSize {
		return fmt.Errorf("page data size mismatch: expected %d bytes, got %d", p.pageSize, len(data))
	}

	// Update cache
	p.cache.Put(id, data)

	// Mark as dirty
	p.dirtyPages[id] = struct{}{}

	// Update dbSize if this is a new page beyond current size
		if uint32(id) > p.dbSize {
		p.dbSize = uint32(id)
	}

	return nil
}

// FlushDirtyPages writes all dirty pages to disk.
func (p *Pager) flushDirtyPagesLocked() error {
	for id := range p.dirtyPages {
		page, ok := p.cache.Get(id)
		if !ok {
			return fmt.Errorf("dirty page %d not found in cache during flush", id)
		}
		offset := int64(id-1) * int64(p.pageSize)
		_, err := p.file.WriteAt(page, offset)
		if err != nil {
			return fmt.Errorf("failed to write dirty page %d to disk: %w", id, err)
		}
	}

	// Sync the file to ensure data is written to persistent storage
	if err := p.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file after flushing dirty pages: %w", err)
	}

	// Clear dirty pages after successful flush
	p.dirtyPages = make(map[pkg.PageID]struct{})

	return nil
}

// FlushDirtyPages writes all dirty pages to disk.
func (p *Pager) FlushDirtyPages() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flushDirtyPagesLocked()
}

// Close closes the underlying file.
func (p *Pager) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Flush any remaining dirty pages before closing
	if err := p.flushDirtyPagesLocked(); err != nil {
		return fmt.Errorf("failed to flush dirty pages before closing pager: %w", err)
	}

	return p.file.Close()
}
