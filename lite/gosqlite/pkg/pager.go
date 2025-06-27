package pkg

import (
    "fmt"
    "io"
    "sort"
    "sync"
)

// Pager is responsible for translating page-IDs (1-indexed) to byte offsets in
// the database file, managing the Adaptive-Replacement-Cache (ARC) and tracking
// dirty pages that must be flushed on commit or Close().  It intentionally does
// NOT understand higher-level database structures – its sole responsibility is
// durable, cache-bounded page access.

type Pager struct {
    vfs  VFS
    file File

    pageSize uint16  // immutable for the lifetime of the Pager instance
    dbSize   uint32  // current database size in pages (lazy-updated)

    cache      *ARCCache          // hot-page cache (ARC)
    dirtyPages map[PageID]struct{} // set of pages modified since last flush

    mu sync.Mutex // protects every field above
}

// NewPager constructs a fully initialised Pager.  The supplied pageSize must
// already have been validated against the SQLite header rules (power-of-two,
// 512-65536).
func NewPager(vfs VFS, file File, pageSize uint16, cachePages int) (*Pager, error) {
    if vfs == nil || file == nil {
        return nil, fmt.Errorf("pager: vfs and file must be non-nil")
    }
    if pageSize < 512 || pageSize > 65536 || (pageSize&(pageSize-1)) != 0 {
        return nil, fmt.Errorf("pager: invalid page size %d", pageSize)
    }
    if cachePages <= 0 {
        cachePages = 256 // sensible default – 256 pages → 1 MiB at 4 KiB pages
    }

    sizeBytes, err := file.Size()
    if err != nil {
        return nil, fmt.Errorf("pager: stat failed: %w", err)
    }

    p := &Pager{
        vfs:        vfs,
        file:       file,
        pageSize:   pageSize,
        dbSize:     uint32(sizeBytes / int64(pageSize)),
        cache:      NewARCCache(cachePages),
        dirtyPages: make(map[PageID]struct{}),
    }

    return p, nil
}

// PageCount returns the current size of the database measured in pages.
func (p *Pager) PageCount() uint32 {
    p.mu.Lock()
    defer p.mu.Unlock()
    return p.dbSize
}

// GetPage retrieves a page, first consulting the ARC cache, otherwise reading
// from disk.  The returned slice is ALWAYS exactly len==pageSize bytes.
func (p *Pager) GetPage(id PageID) (Page, error) {
    if id == 0 {
        return nil, fmt.Errorf("pager: pageID 0 is invalid – pages are 1-indexed")
    }

    p.mu.Lock()
    // fast-path: in-cache → return immediately
    if pg, ok := p.cache.Get(id); ok {
        p.mu.Unlock()
        return pg, nil
    }
    // not cached – we must read from disk; hold reference to file, but release
    // cache lock so ReadAt can run without blocking other readers.
    p.mu.Unlock()

    // allocate outside lock – avoid blocking; we cannot safely reuse the slice
    // because other goroutines may keep references.
    buf := make(Page, p.pageSize)
    offset := int64(id-1) * int64(p.pageSize)
    n, err := p.file.ReadAt(buf, offset)
    if err != nil && err != io.EOF {
        return nil, fmt.Errorf("pager: read page %d failed: %w", id, err)
    }
    if n != int(p.pageSize) {
        // short read → treat as zero-page per SQLite semantics when extending
        for i := n; i < int(p.pageSize); i++ {
            buf[i] = 0
        }
    }

    // store into cache under lock
    p.mu.Lock()
    p.cache.Put(id, buf)
    p.mu.Unlock()

    return buf, nil
}

// WritePage copies the supplied data into the cache and marks the page dirty.
// The caller must supply exactly pageSize bytes.
func (p *Pager) WritePage(id PageID, data Page) error {
    if id == 0 {
        return fmt.Errorf("pager: pageID 0 is invalid – pages are 1-indexed")
    }
    if uint16(len(data)) != p.pageSize {
        return fmt.Errorf("pager: data length %d does not match page size %d", len(data), p.pageSize)
    }

    // Make a copy of the slice – the caller may mutate it after return.
    pageCopy := make(Page, p.pageSize)
    copy(pageCopy, data)

    p.mu.Lock()
    p.cache.Put(id, pageCopy)
    p.dirtyPages[id] = struct{}{}
    if uint32(id) > p.dbSize {
        p.dbSize = uint32(id)
    }
    p.mu.Unlock()
    return nil
}

// FlushDirtyPages persists every dirty page in LRU-order to disk and fsyncs the
// underlying handle.  Callers should hold no locks or slices referencing cached
// pages while invoking FlushDirtyPages().
func (p *Pager) FlushDirtyPages() error {
    p.mu.Lock()
    // build slice of dirty IDs to write in deterministic order (ascending)
    ids := make([]PageID, 0, len(p.dirtyPages))
    for id := range p.dirtyPages {
        ids = append(ids, id)
    }
    p.mu.Unlock()

    // sort ascending for stable ordering – avoids excessive fragmentation
    // across WAL/journal; p.dirtyPages can be large so use sort.Slice.
    sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

    for _, id := range ids {
        p.mu.Lock()
        pg, ok := p.cache.Get(id)
        p.mu.Unlock()
        if !ok {
            return fmt.Errorf("pager: dirty page %d vanished from cache", id)
        }

        offset := int64(id-1) * int64(p.pageSize)
        n, err := p.file.WriteAt(pg, offset)
        if err != nil {
            return fmt.Errorf("pager: write page %d failed: %w", id, err)
        }
        if n != int(p.pageSize) {
            return fmt.Errorf("pager: short write on page %d (wrote %d bytes)", id, n)
        }
    }

    if err := p.file.Sync(); err != nil {
        return fmt.Errorf("pager: fsync failed: %w", err)
    }

    // success – clear dirty map
    p.mu.Lock()
    p.dirtyPages = make(map[PageID]struct{})
    p.mu.Unlock()
    return nil
}

// Close flushes dirty pages and closes the underlying file.
func (p *Pager) Close() error {
    if err := p.FlushDirtyPages(); err != nil {
        return err
    }
    return p.file.Close()
}
