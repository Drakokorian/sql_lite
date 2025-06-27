package pkg

import (
	"container/list"
	"sync"
)

// arcEntry represents an entry in the ARC cache.
type arcEntry struct {
	key   PageID
	value Page
}

// ARCCache implements the Adaptive Replacement Cache (ARC) algorithm.
type ARCCache struct {
	capacity int // Maximum number of pages in the cache

	// T1: pages seen once recently (L1-ARC)
	t1_lru *list.List
	t1_map map[PageID]*list.Element

	// T2: pages seen multiple times recently (L2-ARC)
	t2_lru *list.List
	t2_map map[PageID]*list.Element

	// B1: pages recently evicted from T1 (ghost list for L1-ARC)
	b1_lru *list.List
	b1_map map[PageID]*list.Element

	// B2: pages recently evicted from T2 (ghost list for L2-ARC)
	b2_lru *list.List
	b2_map map[PageID]*list.Element

	// p: target size for T1 (adaptively adjusted)
	p int

	mu sync.Mutex // Mutex to protect concurrent access to the cache
}

// NewARCCache creates a new ARCCache with the given capacity.
func NewARCCache(capacity int) *ARCCache {
	if capacity <= 0 {
		panic("ARC cache capacity must be greater than 0")
	}
	return &ARCCache{
		capacity: capacity,
		t1_lru:   list.New(),
		t1_map:   make(map[PageID]*list.Element),
		t2_lru:   list.New(),
		t2_map:   make(map[PageID]*list.Element),
		b1_lru:   list.New(),
		b1_map:   make(map[PageID]*list.Element),
		b2_lru:   list.New(),
		b2_map:   make(map[PageID]*list.Element),
	}
}

// Get retrieves a page from the cache.
func (c *ARCCache) Get(id PageID) (Page, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.t1_map[id]; ok {
		// Hit in T1, move to T2
		c.t1_lru.Remove(elem)
		delete(c.t1_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		return elem.Value.(*arcEntry).value, true
	} else if elem, ok := c.t2_map[id]; ok {
		// Hit in T2, move to front of T2
		c.t2_lru.MoveToFront(elem)
		return elem.Value.(*arcEntry).value, true
	} else if elem, ok := c.b1_map[id]; ok {
		// Hit in B1, move to T2 and adapt p
		c.p = min(c.capacity, c.p+max(1, c.len(c.b2_lru)/c.len(c.b1_lru)))
		c.replace(false)
		// Move from B1 to T2
		entry := elem.Value.(*arcEntry)
		c.b1_lru.Remove(elem)
		delete(c.b1_map, id)
		c.t2_lru.PushFront(entry)
		c.t2_map[id] = c.t2_lru.Front()
		return entry.value, true
	} else if elem, ok := c.b2_map[id]; ok {
		// Hit in B2, move to T2 and adapt p
		c.p = max(0, c.p-max(1, c.len(c.b1_lru)/c.len(c.b2_lru)))
		c.replace(true)
		// Move from B2 to T2
		entry := elem.Value.(*arcEntry)
		c.b2_lru.Remove(elem)
		delete(c.b2_map, id)
		c.t2_lru.PushFront(entry)
		c.t2_map[id] = c.t2_lru.Front()
		return entry.value, true
	}

	return nil, false
}

// Put adds a page to the cache.
func (c *ARCCache) Put(id PageID, page Page) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &arcEntry{key: id, value: page}

	if elem, ok := c.t1_map[id]; ok {
		// Already in T1, update and move to T2
		c.t1_lru.Remove(elem)
		delete(c.t1_map, id)
		c.t2_lru.PushFront(entry)
		c.t2_map[id] = c.t2_lru.Front()
	} else if elem, ok := c.t2_map[id]; ok {
		// Already in T2, update and move to front of T2
		c.t2_lru.MoveToFront(elem)
		elem.Value = entry // Update entry content
	} else if elem, ok := c.b1_map[id]; ok {
		// In B1, remove from B1 and add to T2
		c.p = min(c.capacity, c.p+max(1, c.len(c.b2_lru)/c.len(c.b1_lru)))
		c.replace(false)
		c.b1_lru.Remove(elem)
		delete(c.b1_map, id)
		c.t2_lru.PushFront(entry)
		c.t2_map[id] = c.t2_lru.Front()
	} else if elem, ok := c.b2_map[id]; ok {
		// In B2, remove from B2 and add to T2
		c.p = max(0, c.p-max(1, c.len(c.b1_lru)/c.len(c.b2_lru)))
		c.replace(true)
		c.b2_lru.Remove(elem)
		delete(c.b2_map, id)
		c.t2_lru.PushFront(entry)
		c.t2_map[id] = c.t2_lru.Front()
	} else {
		// New page
		if c.len(c.t1_lru)+c.len(c.b1_lru) == c.capacity {
			if c.len(c.t1_lru) < c.capacity {
				// B1 is full, move LRU from B1 to B2
				oldest := c.b1_lru.Back()
				delete(c.b1_map, oldest.Value.(*arcEntry).key)
				c.b1_lru.Remove(oldest)
				c.b2_lru.PushFront(oldest)
				c.b2_map[oldest.Value.(*arcEntry).key] = oldest
			}
			// T1 is full, move LRU from T1 to B1
			oldest := c.t1_lru.Back()
			delete(c.t1_map, oldest.Value.(*arcEntry).key)
			c.t1_lru.Remove(oldest)
			c.b1_lru.PushFront(oldest)
			c.b1_map[oldest.Value.(*arcEntry).key] = oldest
		}
		// If T1+T2 is full, replace a page
		if c.len(c.t1_lru)+c.len(c.t2_lru) >= c.capacity {
			c.replace(false) // Replace from T1 or T2
		}
		// Add to T1
		c.t1_lru.PushFront(entry)
		c.t1_map[id] = c.t1_lru.Front()
	}
}

// replace evicts a page from T1 or T2 to make space.
func (c *ARCCache) replace(b2Hit bool) {
	// Case 1: T1 is larger than target p, or T1 is at target p and B2 is not empty
	// Evict from T1 to B1
	if c.len(c.t1_lru) > 0 && ((c.len(c.t1_lru) > c.p) || (c.len(c.t1_lru) == c.p && c.len(c.b2_lru) > 0 && !b2Hit)) {
		oldest := c.t1_lru.Back()
		delete(c.t1_map, oldest.Value.(*arcEntry).key)
		c.t1_lru.Remove(oldest)
		c.b1_lru.PushFront(oldest)
		c.b1_map[oldest.Value.(*arcEntry).key] = oldest
	} else if c.len(c.t2_lru) > 0 {
		// Case 2: Evict from T2 to B2
		oldest := c.t2_lru.Back()
		delete(c.t2_map, oldest.Value.(*arcEntry).key)
		c.t2_lru.Remove(oldest)
		c.b2_lru.PushFront(oldest)
		c.b2_map[oldest.Value.(*arcEntry).key] = oldest
	}
}

// len returns the length of a list.List (helper for clarity).
func (c *ARCCache) len(l *list.List) int {
	return l.Len()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
