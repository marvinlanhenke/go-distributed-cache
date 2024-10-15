package cache_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestCacheSetGet(t *testing.T) {
	cache := cache.New(10)
	cache.Set("key_a", "value_a", time.Hour*1)

	result, ok := cache.Get("key_a")
	require.True(t, ok, "unexpected value, expected %v got %v", true, ok)
	require.Equal(t, result, "value_a", "unexpected value, expected %s got %s", "value_a", result)
}

func TestCacheTTL(t *testing.T) {
	cache := cache.New(10)
	cache.Set("key_a", "value_a", time.Millisecond*1)
	cache.Set("key_b", "value_b", time.Hour*1)

	time.Sleep(time.Millisecond * 2)

	_, ok := cache.Get("key_a")
	require.False(t, ok, "unexpected value, expected %v got %v", false, ok)

	result, ok := cache.Get("key_b")
	require.True(t, ok, "unexpected value, expected %v got %v", true, ok)
	require.Equal(t, result, "value_b", "unexpected value, expected %s got %s", "value_a", result)
}

func TestCacheLRU(t *testing.T) {
	cache := cache.New(2)
	cache.Set("key_a", "value_a", time.Hour*1)
	cache.Set("key_b", "value_b", time.Hour*1)
	cache.Set("key_c", "value_c", time.Hour*1)

	_, ok := cache.Get("key_a")
	require.False(t, ok, "unexpected value, expected %v got %v", false, ok)

	result, ok := cache.Get("key_b")
	require.True(t, ok, "unexpected value, expected %v got %v", true, ok)
	require.Equal(t, result, "value_b", "unexpected value, expected %s got %s", "value_b", result)

	result, ok = cache.Get("key_c")
	require.True(t, ok, "unexpected value, expected %v got %v", true, ok)
	require.Equal(t, result, "value_c", "unexpected value, expected %s got %s", "value_c", result)
}

func TestCacheConcurrency(t *testing.T) {
	cache := cache.New(10)
	var wg sync.WaitGroup

	numGoroutines := 10
	opsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				key := strconv.Itoa(j)
				value := "value" + key

				cache.Set(key, value, time.Hour*1)
				result, ok := cache.Get(key)

				if ok {
					require.Equal(t, value, result, "unexpected value for key %s: got %s, expected %s", key, result, value)
				}

			}
		}()
	}
}
