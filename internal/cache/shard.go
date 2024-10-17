package cache

import (
	"container/list"
	"sync"
	"time"
)

type shard struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	eviction *list.List
	capacity int
}

func (s *shard) evictTTL(item *cacheItem, elem *list.Element, key string) bool {
	if time.Now().After(item.expiryTime) {
		s.eviction.Remove(elem)
		delete(s.items, key)
		return true
	}
	return false
}

func (s *shard) evictLRU() {
	elem := s.eviction.Back()
	if elem != nil {
		s.eviction.Remove(elem)
		entry := elem.Value.(*listEntry)
		delete(s.items, entry.key)
	}
}
