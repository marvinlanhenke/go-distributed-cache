package cache_test

import (
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
