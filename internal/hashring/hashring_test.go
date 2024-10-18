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

func TestHashRingReplication(t *testing.T) {
	hr := hashring.New()

	hr.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	require.Equal(t, 1, hr.Replication, "expected replication to be %d, instead got %d", 1, hr.Replication)

	hr.Add(&hashring.Node{ID: "node2", Addr: "localhost:8081"})
	require.Equal(t, 2, hr.Replication, "expected replication to be %d, instead got %d", 2, hr.Replication)

	hr.Add(&hashring.Node{ID: "node3", Addr: "localhost:8082"})
	require.Equal(t, 2, hr.Replication, "expected replication to be %d, instead got %d", 2, hr.Replication)

	nodes, ok := hr.GetNodes("node1")
	require.True(t, ok, "expected %v, instead got %v", true, ok)
	require.Len(t, nodes, 2, "expected len of %d, instead got %d", 2, len(nodes))

	nodes, ok = hr.GetNodes("node2")
	require.True(t, ok, "expected %v, instead got %v", true, ok)
	require.Len(t, nodes, 2, "expected len of %d, instead got %d", 2, len(nodes))

	nodes, ok = hr.GetNodes("node3")
	require.True(t, ok, "expected %v, instead got %v", true, ok)
	require.Len(t, nodes, 2, "expected len of %d, instead got %d", 2, len(nodes))
}

func TestHashRingReplicationWithEmptyRing(t *testing.T) {
	hr := hashring.New()

	nodes, ok := hr.GetNodes("node1")
	require.False(t, ok, "expected %v, instead got %v", false, ok)
	require.Nil(t, nodes, "expected nodes to be nil, instead got %v", nodes)
}

func TestHashRingReplicationWithGreaterReplicationFactor(t *testing.T) {
	hr := hashring.New()
	hr.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr.Replication = 2

	nodes, ok := hr.GetNodes("node1")
	require.False(t, ok, "expected %v, instead got %v", false, ok)
	require.Nil(t, nodes, "expected nodes to be nil, instead got %v", nodes)
}

func TestHashRingReplicationWithDuplicates(t *testing.T) {
	hr := hashring.New()
	hr.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr.Add(&hashring.Node{ID: "node1", Addr: "localhost:8080"})
	hr.Add(&hashring.Node{ID: "node2", Addr: "localhost:8081"})
	hr.Add(&hashring.Node{ID: "node2", Addr: "localhost:8081"})

	nodes, ok := hr.GetNodes("node1")
	require.False(t, ok, "expected %v, instead got %v", false, ok)
	require.Nil(t, nodes, "expected nodes to be nil, instead got %v", nodes)
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
