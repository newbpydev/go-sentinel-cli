// Package api provides the core API functionality for the Go Sentinel service.
// It includes caching, request handling, and integration with the core test engine.
package api

import (
	"container/list"
	"sync"
)

// ResultCache provides a thread-safe cache for storing test results
// with a fixed capacity using a least recently used (LRU) eviction policy.
type ResultCache struct {
	capacity int
	mu       sync.Mutex
	items    map[string]*list.Element
	order    *list.List
}

type cacheEntry struct {
	key   string
	value interface{}
}

// NewResultCache creates a new ResultCache with the specified capacity.
// The capacity determines how many items can be stored before old ones are evicted.
func NewResultCache(capacity int) *ResultCache {
	return &ResultCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Set adds or updates a value in the cache with the given key.
// If the cache is at capacity, the least recently used item is evicted.
func (c *ResultCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			entry := oldest.Value.(*cacheEntry)
			delete(c.items, entry.key)
			c.order.Remove(oldest)
		}
	}
	elem := c.order.PushFront(&cacheEntry{key, value})
	c.items[key] = elem
}

// Get retrieves a value from the cache by key.
// Returns the value and a boolean indicating whether the key was found.
func (c *ResultCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}
