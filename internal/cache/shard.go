package cache

import (
	"container/list"
	"sync"
	"time"
)

// Represents a partition of the cache that stores a subset of cache items.
type shard struct {
	mu       sync.RWMutex             // Mutex for synchronizing read and write access to the shard.
	items    map[string]*list.Element // Map for fast lookup of cache items by key.
	eviction *list.List               // Doubly linked list to track item usage for LRU eviction.
	capacity int                      // Maximum number of items the shard can hold before eviction is triggered.
}

// Checks if a cache item has expired based on its TTL (time-to-live).
//
// If the item is expired, it is removed from both the eviction list and the items map.
// Returns true if the item was evicted, false otherwise.
func (s *shard) evictTTL(item *cacheItem, elem *list.Element, key string) bool {
	if time.Now().After(item.expiryTime) {
		s.eviction.Remove(elem)
		delete(s.items, key)
		return true
	}
	return false
}

// Evicts the least-recently-used (LRU) item from the shard when the capacity is exceeded.
// The item at the back of the eviction list (the least recently used) is removed from both the list and the items map.
func (s *shard) evictLRU() {
	elem := s.eviction.Back()
	if elem != nil {
		s.eviction.Remove(elem)
		entry := elem.Value.(*listEntry)
		delete(s.items, entry.key)
	}
}
