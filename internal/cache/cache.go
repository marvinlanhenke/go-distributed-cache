package cache

import (
	"container/list"
	"sync"
	"time"
)

type cacheItem struct {
	value      string
	expiryTime time.Time
}

type entry struct {
	key  string
	item cacheItem
}

type Cache struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	eviction *list.List
	capacity int
}

func New(capacity int) *Cache {
	return &Cache{
		items:    make(map[string]*list.Element),
		eviction: list.New(),
		capacity: capacity,
	}
}

func (c *Cache) Set(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.eviction.Remove(elem)
		delete(c.items, key)
	}

	if c.eviction.Len() >= c.capacity {
		c.evictLRU()
	}

	item := cacheItem{
		value:      value,
		expiryTime: time.Now().Add(ttl),
	}
	elem := c.eviction.PushFront(&entry{key, item})
	c.items[key] = elem
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.items[key]
	if !ok {
		return "", false
	}

	item := elem.Value.(*entry).item
	if time.Now().After(item.expiryTime) {
		c.eviction.Remove(elem)
		delete(c.items, key)
		return "", false
	}

	c.eviction.MoveToFront(elem)

	return item.value, true
}

func (c *Cache) evictLRU() {
	elem := c.eviction.Back()
	if elem != nil {
		c.eviction.Remove(elem)
		entry := elem.Value.(*entry)
		delete(c.items, entry.key)
	}
}
