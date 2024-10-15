package hashring_test

import (
	"testing"

	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/stretchr/testify/require"
)

func TestHashRingAddRemoveNode(t *testing.T) {
	hashRing := hashring.New()
	node := &hashring.Node{ID: "node1", Addr: "localhost:8080"}

	hashRing.Add(node)
	require.GreaterOrEqual(t, hashRing.Size(), 1, "expected size of hashring to be >= 1, instead size is %d", hashRing.Size())

	hashRing.Remove("node1")
	require.Equal(t, hashRing.Size(), 0, "expected size of hashring to be zero, instead size is %d", hashRing.Size())
}

func TestHashRingGetNode(t *testing.T) {
	hashRing := hashring.New()
	node := &hashring.Node{ID: "node1", Addr: "localhost:8080"}
	hashRing.Add(node)

	result := hashRing.Get("node1")
	require.Equal(t, node, result, "expected to get node %v, instead got %v", node, result)
}

func TestHashRingGetNodeWithEmptyRing(t *testing.T) {
	hashRing := hashring.New()

	result := hashRing.Get("node1")
	require.Nil(t, result, "expected result to be nil, instead got %v", result)
}
