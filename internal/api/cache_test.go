package api

import (
	"sync"
	"testing"
)

func TestCacheStoresRecentResults(t *testing.T) {
	cache := NewResultCache(3)
	cache.Set("a", "resultA")
	cache.Set("b", "resultB")
	cache.Set("c", "resultC")
	if v, ok := cache.Get("a"); !ok || v != "resultA" {
		t.Errorf("expected to find 'resultA', got %v", v)
	}
}

func TestCacheEvictsOldest(t *testing.T) {
	cache := NewResultCache(2)
	cache.Set("a", "resultA")
	cache.Set("b", "resultB")
	cache.Set("c", "resultC") // should evict "a"
	if _, ok := cache.Get("a"); ok {
		t.Error("expected 'a' to be evicted")
	}
	if v, ok := cache.Get("b"); !ok || v != "resultB" {
		t.Errorf("expected to find 'resultB', got %v", v)
	}
	if v, ok := cache.Get("c"); !ok || v != "resultC" {
		t.Errorf("expected to find 'resultC', got %v", v)
	}
}

func TestCacheThreadSafety(_ *testing.T) {
	cache := NewResultCache(10)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(string(rune('a'+i%10)), i)
			cache.Get(string(rune('a' + i%10)))
		}(i)
	}
	wg.Wait()
}
