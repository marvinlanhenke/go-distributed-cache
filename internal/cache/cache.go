package cache

import (
	"container/list"
	"hash/fnv"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
)

// Represents an individual cache entry.
type cacheItem struct {
	value      string    // The actual cached value.
	version    int       // Version of the cache item, used to manage updates.
	expiryTime time.Time // Time when the cache item will expire.
}

// Represents an entry in the eviction list.
type listEntry struct {
	key  string     // The key associated with the cache item.
	item *cacheItem // Pointer to the actual cache item.
}

// Cache represents a distributed cache with multiple shards for concurrency and efficiency.
// Each shard manages a subset of cache entries to reduce contention.
type Cache struct {
	shards    []*shard      // Slice of cache shards.
	numShards int           // Number of shards for distributing cache keys.
	ttl       time.Duration // Time-to-live for cache entries.
}

// Initializes and returns a new `Cache` instance.
// It distributes the capacity evenly across all shards.
func New(numShards, capacity int, ttl time.Duration) *Cache {
	shards := make([]*shard, numShards)
	capacityPerShard := capacity / numShards
	for i := 0; i < numShards; i++ {
		shards[i] = &shard{
			items:    make(map[string]*list.Element),
			eviction: list.New(),
			capacity: capacityPerShard,
		}
	}
	return &Cache{
		shards:    shards,
		numShards: numShards,
		ttl:       ttl,
	}
}

// Set adds or updates a cache entry with the specified key and value from the SetRequest.
// If the cache exceeds its capacity, the least-recently-used (LRU) item is evicted.
func (c *Cache) Set(req *pb.SetRequest) {
	shard := c.getShard(req.Key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	var nextVersion int = 0
	if elem, ok := shard.items[req.Key]; ok {
		nextVersion = elem.Value.(*listEntry).item.version + 1
		shard.eviction.Remove(elem)
		delete(shard.items, req.Key)
	}

	if shard.eviction.Len() >= shard.capacity {
		shard.evictLRU()
	}

	item := &cacheItem{
		value:      req.Value,
		version:    nextVersion,
		expiryTime: time.Now().Add(c.ttl),
	}
	elem := shard.eviction.PushFront(&listEntry{key: req.Key, item: item})
	shard.items[req.Key] = elem
}

// Retrieves a cache entry by key and returns a GetResponse if the key exists and has not expired.
// If the item is found, it is moved to the front of the eviction list to mark it as recently used.
func (c *Cache) Get(req *pb.GetRequest) (*pb.GetResponse, bool) {
	shard := c.getShard(req.Key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	elem, ok := shard.items[req.Key]
	if !ok {
		return nil, false
	}

	item := elem.Value.(*listEntry).item
	if shard.evictTTL(item, elem, req.Key) {
		return nil, false
	}

	shard.eviction.MoveToFront(elem)

	return &pb.GetResponse{
		Value:   item.value,
		Version: uint32(item.version),
	}, true
}

// Determines the appropriate shard for a given cache key by hashing the key.
func (c *Cache) getShard(key string) *shard {
	hash := fnv32(key)
	return c.shards[hash%uint32(c.numShards)]
}

// Hashes a string key using the FNV-1a hash algorithm.
func fnv32(key string) uint32 {
	hsh := fnv.New32a()
	hsh.Write([]byte(key))
	return hsh.Sum32()
}
