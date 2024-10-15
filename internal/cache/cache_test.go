package cache_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestCacheSetGet(t *testing.T) {
	cache := cache.New()

	cache.Set("key_a", "value_a")
	value, ok := cache.Get("key_a")

	require.True(t, ok, "failed to get a value")
	require.Equal(t, value, "value_a", "failed to get the correct value")
}

func TestCacheConcurrency(t *testing.T) {
	cache := cache.New()
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

				cache.Set(key, value)
				result, ok := cache.Get(key)

				if ok {
					require.Equal(t, value, result, "unexpected value for key %s: got %s, expected %s", key, result, value)
				}

			}
		}()
	}
}
