package hashring

import (
	"crypto/sha1"
	"encoding/binary"
	"sort"
	"sync"
)

// Represents a node in the hash ring, identified by its unique ID and associated with an address.
type Node struct {
	ID   string // Unique identifier of the node.
	Addr string // Network address of the node.
}

// Represents a node in the hash ring, along with its hashed key value.
type member struct {
	hash uint32 // Hash of the node's ID.
	node *Node  // Pointer to the actual node.
}

// HashRing represents a consistent hash ring used for distributing keys across nodes.
// It supports adding, removing, and retrieving nodes based on the hash of a key.
type HashRing struct {
	mu          sync.Mutex // Mutex to ensure thread-safe operations on the ring.
	members     []member   // Slice of members (nodes) in the hash ring.
	Replication int        // Number of nodes to replicate each key to.
}

// Creates and returns an empty HashRing instance.
func New() *HashRing {
	return &HashRing{}
}

// Returns the number of members currently in the hash ring.
func (hr *HashRing) Size() int {
	return len(hr.members)
}

// Checks if the hash ring has no members and returns true if empty, false otherwise.
func (hr *HashRing) IsEmpty() bool {
	return hr.Size() == 0
}

// Adds a new node to the hash ring, hashing the node's ID and inserting it into the sorted list of members.
// The replication factor is updated after the node is added.
func (hr *HashRing) Add(node *Node) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hash := hr.hash(node.ID)
	member := member{hash: hash, node: node}
	hr.members = append(hr.members, member)

	sort.Slice(hr.members, func(i, j int) bool {
		return hr.members[i].hash < hr.members[j].hash
	})

	hr.Replication = hr.Size()/2 + 1
}

// Removes a node from the hash ring by its ID, adjusting the list of members accordingly.
func (hr *HashRing) Remove(nodeID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for i, member := range hr.members {
		if member.node.ID == nodeID {
			hr.members = append(hr.members[:i], hr.members[i+1:]...)
			break
		}
	}
}

// Returns a list of nodes that should be responsible for the given key based on its hash.
// The number of nodes returned is determined by the replication factor. If enough nodes cannot be found, it returns false.
func (hr *HashRing) GetNodes(key string) ([]*Node, bool) {
	if hr.IsEmpty() || hr.Replication > hr.Size() {
		return nil, false
	}

	hr.mu.Lock()
	defer hr.mu.Unlock()

	hash := hr.hash(key)
	index := sort.Search(hr.Size(), func(i int) bool {
		return hr.members[i].hash >= hash
	})

	if index == hr.Size() {
		index = 0
	}

	seen := make(map[string]struct{})
	nodes := make([]*Node, 0, hr.Replication)

	currentIndex := index
	for len(nodes) < hr.Replication {
		node := hr.members[currentIndex].node
		if _, exists := seen[node.ID]; !exists {
			nodes = append(nodes, node)
			seen[node.ID] = struct{}{}
		}

		currentIndex++
		if currentIndex >= hr.Size() {
			currentIndex = 0
		}

		if currentIndex == index {
			break
		}
	}

	if len(nodes) < hr.Replication {
		return nil, false
	}

	return nodes, true
}

// Computes a 32-bit hash of a given key using the SHA-1 hashing algorithm.
// The first 4 bytes of the SHA-1 hash are used to generate the 32-bit hash value.
func (hr *HashRing) hash(key string) uint32 {
	sum := sha1.Sum([]byte(key))
	return binary.LittleEndian.Uint32(sum[:4])
}
