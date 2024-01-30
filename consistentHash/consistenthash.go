package consistentHash

import (
	"slices"
	"sort"
)

const maxNodes int = 1337

func New(hasher Hasher) *ConsistentHash {
	return &ConsistentHash{
		maxNodes: maxNodes,
		hasher:   hasher,
	}
}

type Hasher interface {
	Hash(key string) int
}
type Node struct {
	name string
	host string
}

type ConsistentHash struct {
	nodes    []Node
	keys     []int
	maxNodes int
	hasher   Hasher
}

func (h *ConsistentHash) hash(key string) int {
	return h.hasher.Hash(key) % h.maxNodes
}

func (h *ConsistentHash) AddNode(n Node) int {
	key := h.hash(n.host)
	index := bisect(h.keys, key)
	h.nodes = slices.Insert(h.nodes, index, n)
	h.keys = slices.Insert(h.keys, index, key)

	return key
}

func (h *ConsistentHash) RemoveNode() {
	// TODO
}

func (h *ConsistentHash) PickNode(key string) Node {
	k := h.hash(key)
	index := bisectRight(h.keys, k)
	return h.nodes[index]
}

func bisectRight(a []int, value int) int {
	return sort.Search(len(a), func(i int) bool { return a[i] > value })
}

func bisect(a []int, value int) int {
	return sort.Search(len(a), func(i int) bool { return a[i] == value })
}
