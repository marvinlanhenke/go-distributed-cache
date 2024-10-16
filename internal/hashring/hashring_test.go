package hashring_test

import (
	"testing"

	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/stretchr/testify/require"
)

func TestHashRingAddGetNode(t *testing.T) {
	hr := hashring.New()
	node := &hashring.Node{ID: "node1", Addr: "localhost:8080"}

	hr.Add(node)
	require.GreaterOrEqual(t, hr.Size(), 1, "expected size >= 1, instead got %v", hr.Size())

	result, _ := hr.Get("node1")
	require.Equal(t, node, result, "expected to retrieve node %v, instead got %v", node, result)
}

func TestHashRingWithEqualNodes(t *testing.T) {
	hr1 := hashring.New()
	hr1.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr1.Add(&hashring.Node{ID: "node2", Addr: "localhost:8081"})

	hr2 := hashring.New()
	hr2.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr2.Add(&hashring.Node{ID: "node2", Addr: "localhost:8081"})

	require.Equal(t, hr1, hr2, "expected both hashrings to be equal")
}

func TestHashRingGetNodeWithEmptyRing(t *testing.T) {
	hr := hashring.New()
	result, _ := hr.Get("node1")

	require.Nil(t, result, "expected result to be nil, instead got %v", result)
}
