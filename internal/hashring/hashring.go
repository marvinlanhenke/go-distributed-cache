package hashring

import (
	"crypto/sha1"
	"encoding/binary"
	"sort"
	"sync"
)

type Node struct {
	ID   string
	Addr string
}

type member struct {
	hash uint32
	node *Node
}

type HashRing struct {
	mu          sync.Mutex
	members     []member
	Replication int
}

func New() *HashRing {
	return &HashRing{}
}

func (hr *HashRing) Size() int {
	return len(hr.members)
}

func (hr *HashRing) IsEmpty() bool {
	return hr.Size() == 0
}

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

func (hr *HashRing) hash(key string) uint32 {
	sum := sha1.Sum([]byte(key))
	return binary.LittleEndian.Uint32(sum[:4])
}
