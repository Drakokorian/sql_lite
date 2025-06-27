package pkg

import "testing"

func TestARCCacheBasic(t *testing.T) {
	cache := NewARCCache(2)

	p1 := Page{1}
	p2 := Page{2}
	p3 := Page{3}

	cache.Put(1, p1)
	cache.Put(2, p2)

	if _, ok := cache.Get(1); !ok {
		t.Error("expected page 1 in cache")
	}

	// Add third page – should trigger eviction due to capacity 2
	cache.Put(3, p3)
	if _, ok := cache.Get(2); ok && cache.len(cache.t1_lru)+cache.len(cache.t2_lru) > 2 {
		t.Error("cache exceeded capacity")
	}
}

