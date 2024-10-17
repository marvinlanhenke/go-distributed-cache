package cache

import (
	"container/list"
	"hash/fnv"
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

type cacheShard struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	eviction *list.List
	capacity int
}

func (cs *cacheShard) evictTTL(item *cacheItem, elem *list.Element, key string) bool {
	if time.Now().After(item.expiryTime) {
		cs.eviction.Remove(elem)
		delete(cs.items, key)
		return true
	}
	return false
}

func (cs *cacheShard) evictLRU() {
	elem := cs.eviction.Back()
	if elem != nil {
		cs.eviction.Remove(elem)
		entry := elem.Value.(*listEntry)
		delete(cs.items, entry.key)
	}
}

type Cache struct {
	shards    []*cacheShard
	numShards int
	ttl       time.Duration
}

func New(numShards, capacity int, ttl time.Duration) *Cache {
	shards := make([]*cacheShard, numShards)
	capacityPerShard := capacity / numShards
	for i := 0; i < numShards; i++ {
		shards[i] = &cacheShard{
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

func (c *Cache) Set(req *pb.SetRequest) {
	shard := c.getShard(req.Key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if elem, ok := shard.items[req.Key]; ok {
		shard.eviction.Remove(elem)
		delete(shard.items, req.Key)
	}

	if shard.eviction.Len() >= shard.capacity {
		shard.evictLRU()
	}

	item := &cacheItem{
		value:      req.Value,
		expiryTime: time.Now().Add(c.ttl),
	}
	elem := shard.eviction.PushFront(&listEntry{key: req.Key, item: item})
	shard.items[req.Key] = elem
}

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
		Value: item.value,
	}, true
}

func (c *Cache) getShard(key string) *cacheShard {
	hash := fnv32(key)
	return c.shards[hash%uint32(c.numShards)]
}

func fnv32(key string) uint32 {
	hsh := fnv.New32a()
	hsh.Write([]byte(key))
	return hsh.Sum32()
}
