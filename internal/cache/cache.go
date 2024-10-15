package cache

type CacheItem struct {
	Value string
}

type Cache struct {
	items map[string]CacheItem
}

func New() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

func (c *Cache) Set(key, value string) {
	c.items[key] = CacheItem{Value: value}
}

func (c *Cache) Get(key string) (string, bool) {
	item, ok := c.items[key]
	if !ok {
		return "", false
	}
	return item.Value, true
}
