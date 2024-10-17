package cache_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/stretchr/testify/require"
)

func TestCacheSetGet(t *testing.T) {
	cache := cache.New(1, 10, 3600*time.Second)
	req := &pb.SetRequest{Key: "key1", Value: "value1"}
	cache.Set(req)

	expected := &pb.GetResponse{Value: "value1"}
	result, ok := cache.Get(&pb.GetRequest{Key: "key1"})
	require.True(t, ok, "unexpected value, expected %v instead got %v", true, ok)
	require.Equal(t, expected, result, "unexpected value, expected %v instead got %v", expected, result)
}

func TestCacheTTLEvicted(t *testing.T) {
	cache := cache.New(1, 10, 1*time.Millisecond)
	cache.Set(&pb.SetRequest{Key: "key1", Value: "value1"})

	time.Sleep(2 * time.Millisecond)

	_, ok := cache.Get(&pb.GetRequest{Key: "key1"})
	require.False(t, ok, "unexpected value, expected %v insteag got %v", false, ok)
}

func TestCacheTTLNotEvicted(t *testing.T) {
	cache := cache.New(1, 10, 10*time.Second)
	cache.Set(&pb.SetRequest{Key: "key1", Value: "value1"})

	_, ok := cache.Get(&pb.GetRequest{Key: "key1"})
	require.True(t, ok, "unexpected value, expected %v insteag got %v", true, ok)
}

func TestCacheLRU(t *testing.T) {
	cache := cache.New(1, 2, 10*time.Second)
	cache.Set(&pb.SetRequest{Key: "key1", Value: "value1"})
	cache.Set(&pb.SetRequest{Key: "key2", Value: "value2"})
	cache.Set(&pb.SetRequest{Key: "key3", Value: "value3"})

	_, ok := cache.Get(&pb.GetRequest{Key: "key1"})
	require.False(t, ok, "unexpected value, expected %v instead got %v", false, ok)

	expected := &pb.GetResponse{Value: "value2"}
	result, ok := cache.Get(&pb.GetRequest{Key: "key2"})
	require.True(t, ok, "unexpected value, expected %v instead got %v", true, ok)
	require.Equal(t, expected, result, "unexpected value, expected %v instead got %v", expected, result)

	expected = &pb.GetResponse{Value: "value3"}
	result, ok = cache.Get(&pb.GetRequest{Key: "key3"})
	require.True(t, ok, "unexpected value, expected %v instead got %v", true, ok)
	require.Equal(t, expected, result, "unexpected value, expected %v instead got %v", expected, result)
}

func TestCacheConcurrency(t *testing.T) {
	cache := cache.New(1, 10, 1*time.Hour)
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

				cache.Set(&pb.SetRequest{Key: key, Value: value})
				result, ok := cache.Get(&pb.GetRequest{Key: key})

				if ok {
					expected := &pb.GetResponse{Value: value}
					require.Equal(t, expected, result, "unexpected value, expected %v instead got %v", expected, result)
				}
			}
		}()
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := cache.New(100, 1000000, time.Second*3600)
	req := &pb.SetRequest{Key: "test-key", Value: "test-value"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(req)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cacheSize := 1000000
	cache := cache.New(100, cacheSize, time.Second*3600)
	var keys []string

	for i := 0; i < cacheSize; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		value := fmt.Sprintf("test-value-%d", i)
		req := &pb.SetRequest{Key: key, Value: value}
		cache.Set(req)
		keys = append(keys, key)
	}

	getReqs := make([]*pb.GetRequest, b.N)
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		getReqs[i] = &pb.GetRequest{Key: key}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get(getReqs[i])
	}
}

func BenchmarkCacheMixedParallel(b *testing.B) {
	cacheSize := 1000000
	cache := cache.New(100, cacheSize, time.Second*3600)
	var keys []string

	for i := 0; i < cacheSize; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		value := fmt.Sprintf("test-value-%d", i)
		req := &pb.SetRequest{Key: key, Value: value}
		cache.Set(req)
		keys = append(keys, key)
	}

	getReqs := make([]*pb.GetRequest, b.N)
	setReqs := make([]*pb.SetRequest, b.N)
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		getReqs[i] = &pb.GetRequest{Key: key}
		setReqs[i] = &pb.SetRequest{Key: key, Value: "value"}
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for p.Next() {
			if rng.Intn(10) < 8 {
				getReq := getReqs[rng.Intn(len(getReqs))]
				cache.Get(getReq)
			} else {
				setReq := setReqs[rng.Intn(len(setReqs))]
				cache.Set(setReq)
			}
		}
	})
}
