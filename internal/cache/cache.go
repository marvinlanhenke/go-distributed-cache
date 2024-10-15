package cache

import "sync"

type CacheItem struct {
	Value string
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

func New() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{Value: value}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return "", false
	}
	return item.Value, true
}
