package pkg

import (
	"container/list"
	"sync"
)

// ARCCache implements the Adaptive Replacement Cache (ARC) algorithm.
type ARCCache struct {
	capacity int
	// T1: pages recently used once
	t1_lru *list.List
	t1_map map[PageID]*list.Element
	// T2: pages used at least twice
	t2_lru *list.List
	t2_map map[PageID]*list.Element
	// B1: pages recently evicted from T1
	b1_lru *list.List
	b1_map map[PageID]*list.Element
	// B2: pages recently evicted from T2
	b2_lru *list.List
	b2_map map[PageID]*list.Element
	// p: target size for T1
	p int
	mu sync.Mutex
}

// NewARCCache creates a new ARCCache with the given capacity.
func NewARCCache(capacity int) *ARCCache {
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
		return elem.Value.(Page), true
	} else if elem, ok := c.t2_map[id]; ok {
		// Hit in T2, move to front of T2
		c.t2_lru.MoveToFront(elem)
		return elem.Value.(Page), true
	} else if elem, ok := c.b1_map[id]; ok {
		// Hit in B1, move to T2 and adapt p
		c.p = min(c.capacity, c.p+max(1, len(c.b2_lru)/len(c.b1_lru)))
		c.replace()
		c.b1_lru.Remove(elem)
		delete(c.b1_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		return elem.Value.(Page), true
	} else if elem, ok := c.b2_map[id]; ok {
		// Hit in B2, move to T2 and adapt p
		c.p = max(0, c.p-max(1, len(c.b1_lru)/len(c.b2_lru)))
		c.replace()
		c.b2_lru.Remove(elem)
		delete(c.b2_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		return elem.Value.(Page), true
	}

	return nil, false
}

// Put adds a page to the cache.
func (c *ARCCache) Put(id PageID, page Page) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.t1_map[id]; ok {
		// Already in T1, update and move to T2
		c.t1_lru.Remove(elem)
		delete(c.t1_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		elem.Value = page // Update page content
	} else if elem, ok := c.t2_map[id]; ok {
		// Already in T2, update and move to front of T2
		c.t2_lru.MoveToFront(elem)
		elem.Value = page // Update page content
	} else if elem, ok := c.b1_map[id]; ok {
		// In B1, remove from B1 and add to T2
		c.p = min(c.capacity, c.p+max(1, len(c.b2_lru)/len(c.b1_lru)))
		c.replace()
		c.b1_lru.Remove(elem)
		delete(c.b1_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		elem.Value = page // Update page content
	} else if elem, ok := c.b2_map[id]; ok {
		// In B2, remove from B2 and add to T2
		c.p = max(0, c.p-max(1, len(c.b1_lru)/len(c.b2_lru)))
		c.replace()
		c.b2_lru.Remove(elem)
		delete(c.b2_map, id)
		c.t2_lru.PushFront(elem)
		c.t2_map[id] = elem
		elem.Value = page // Update page content
	} else {
		// New page
		if len(c.t1_lru)+len(c.b1_lru) == c.capacity {
			if len(c.t1_lru) < c.capacity {
				// B1 is full, move LRU from B1 to B2
				oldest := c.b1_lru.Back()
				delete(c.b1_map, oldest.Value.(PageID))
				c.b1_lru.Remove(oldest)
				c.b2_lru.PushFront(oldest)
				c.b2_map[oldest.Value.(PageID)] = oldest
			}
			// T1 is full, move LRU from T1 to B1
			oldest := c.t1_lru.Back()
			delete(c.t1_map, oldest.Value.(PageID))
			c.t1_lru.Remove(oldest)
			c.b1_lru.PushFront(oldest)
			c.b1_map[oldest.Value.(PageID)] = oldest
		}
		// If T1+T2 is full, replace a page
		if len(c.t1_lru)+len(c.t2_lru) >= c.capacity {
			c.replace()
		}
		// Add to T1
		newElem := c.t1_lru.PushFront(page)
		c.t1_map[id] = newElem
	}
}

// replace evicts a page from T1 or T2 to make space.
func (c *ARCCache) replace() {
	if len(c.t1_lru) > 0 && ((len(c.t1_lru) > c.p) || (len(c.t1_lru) == c.p && len(c.b2_lru) > 0)) {
		// Evict from T1 to B1
		oldest := c.t1_lru.Back()
		delete(c.t1_map, oldest.Value.(PageID))
		c.t1_lru.Remove(oldest)
		c.b1_lru.PushFront(oldest)
		c.b1_map[oldest.Value.(PageID)] = oldest
	} else {
		// Evict from T2 to B2
		oldest := c.t2_lru.Back()
		delete(c.t2_map, oldest.Value.(PageID))
		c.t2_lru.Remove(oldest)
		c.b2_lru.PushFront(oldest)
		c.b2_map[oldest.Value.(PageID)] = oldest
	}
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
