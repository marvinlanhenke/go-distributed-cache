package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
)

type cacheItem struct {
	value      string
	expiryTime time.Time
}

type listEntry struct {
	key  string
	item *cacheItem
}

type Cache struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	eviction *list.List
	capacity int
	ttl      time.Duration
}

func New(capacity int, ttl time.Duration) *Cache {
	return &Cache{
		items:    make(map[string]*list.Element),
		eviction: list.New(),
		capacity: capacity,
		ttl:      ttl,
	}
}

func (c *Cache) Set(req *pb.SetRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[req.Key]; ok {
		c.eviction.Remove(elem)
		delete(c.items, req.Key)
	}

	if c.eviction.Len() >= c.capacity {
		c.evictLRU()
	}

	item := &cacheItem{
		value:      req.Value,
		expiryTime: time.Now().Add(c.ttl),
	}
	elem := c.eviction.PushFront(&listEntry{key: req.Key, item: item})
	c.items[req.Key] = elem
}

func (c *Cache) Get(req *pb.GetRequest) (*pb.GetResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.items[req.Key]
	if !ok {
		return nil, false
	}

	item := elem.Value.(*listEntry).item
	if c.evictTTL(item, elem, req.Key) {
		return nil, false
	}

	c.eviction.MoveToFront(elem)

	return &pb.GetResponse{
		Value: item.value,
	}, true
}

func (c *Cache) evictTTL(item *cacheItem, elem *list.Element, key string) bool {
	if time.Now().After(item.expiryTime) {
		c.eviction.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

func (c *Cache) evictLRU() {
	elem := c.eviction.Back()
	if elem != nil {
		c.eviction.Remove(elem)
		entry := elem.Value.(*listEntry)
		delete(c.items, entry.key)
	}
}
