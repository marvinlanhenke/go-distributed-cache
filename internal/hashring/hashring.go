package hashring

import (
	"crypto/sha1"
	"sort"
	"sync"
)

type Node struct {
	ID   string
	Addr string
}

type hashNode struct {
	hash uint32
	node *Node
}

type HashRing struct {
	mu     sync.Mutex
	hashes []hashNode
}

func New() *HashRing {
	return &HashRing{}
}

func (h *HashRing) Size() int {
	return len(h.hashes)
}

func (h *HashRing) IsEmpty() bool {
	return h.Size() == 0
}

func (h *HashRing) Add(node *Node) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hash := h.hash(node.ID)
	hashNode := hashNode{hash: hash, node: node}
	h.hashes = append(h.hashes, hashNode)

	sort.Slice(h.hashes, func(i, j int) bool {
		return h.hashes[i].hash < h.hashes[j].hash
	})
}

func (h *HashRing) Remove(nodeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, hn := range h.hashes {
		if hn.node.ID == nodeID {
			h.hashes = append(h.hashes[:i], h.hashes[i+1:]...)
			break
		}
	}
}

func (h *HashRing) Get(key string) *Node {
	if h.IsEmpty() {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	hash := h.hash(key)
	idx := sort.Search(len(h.hashes), func(i int) bool {
		return h.hashes[i].hash >= hash
	})

	if idx == len(h.hashes) {
		idx = 0
	}

	return h.hashes[idx].node
}

func (h *HashRing) hash(key string) uint32 {
	hsh := sha1.New()
	hsh.Write([]byte(key))
	return h.bytesToUint32(hsh.Sum(nil))
}

func (h *HashRing) bytesToUint32(b []byte) uint32 {
	return (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | (uint32(b[3]))
}
