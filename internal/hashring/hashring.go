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
	mu      sync.Mutex
	members []member
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
}

func (hr *HashRing) Get(nodeID string) (*Node, bool) {
	if hr.IsEmpty() {
		return nil, false
	}

	hr.mu.Lock()
	defer hr.mu.Unlock()

	hash := hr.hash(nodeID)
	index := sort.Search(hr.Size(), func(i int) bool {
		return hr.members[i].hash >= hash
	})

	if index == hr.Size() {
		index = 0
	}

	return hr.members[index].node, true
}

func (hr *HashRing) hash(nodeID string) uint32 {
	sum := sha1.Sum([]byte(nodeID))
	return binary.LittleEndian.Uint32(sum[:4])
}
