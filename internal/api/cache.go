package api

import (
	"container/list"
	"sync"
)

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

func NewResultCache(capacity int) *ResultCache {
	return &ResultCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

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

func (c *ResultCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}
