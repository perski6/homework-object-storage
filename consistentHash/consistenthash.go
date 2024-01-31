package consistentHash

import (
	"sort"
	"sync"
)

const maxNodes int = 1337

// NodeProvider is an interface that defines methods to interact with a consistent hash ring.
type NodeProvider[T any] interface {
	PickNode(key string) Node[T]
	AddNode(n Node[T]) int
	RemoveNode(host string)
}

func New[T any](hasher Hasher) *ConsistentHash[T] {
	return &ConsistentHash[T]{
		maxNodes: maxNodes,
		hasher:   hasher,
	}
}

type Hasher interface {
	Hash(key string) int
}

// Node has a generic type T which can be used to store any type of client or data.
type Node[T any] struct {
	Name   string
	Client T
}

type ConsistentHash[T any] struct {
	nodes    []Node[T]
	keys     []int
	maxNodes int
	hasher   Hasher
	mu       sync.RWMutex
}

func (h *ConsistentHash[T]) hash(key string) int {
	return h.hasher.Hash(key) % h.maxNodes
}

func (h *ConsistentHash[T]) AddNode(n Node[T]) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.hash(n.Name)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] >= key })
	if index < len(h.keys) && h.keys[index] == key {
		// Key already exists, do not add again
		return key
	}
	h.nodes = append(h.nodes[:index], append([]Node[T]{n}, h.nodes[index:]...)...)
	h.keys = append(h.keys[:index], append([]int{key}, h.keys[index:]...)...)

	return key
}

func (h *ConsistentHash[T]) RemoveNode(host string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.hash(host)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] >= key })
	if index < len(h.keys) && h.keys[index] == key {
		h.nodes = append(h.nodes[:index], h.nodes[index+1:]...)
		h.keys = append(h.keys[:index], h.keys[index+1:]...)
	}
}

func (h *ConsistentHash[T]) PickNode(key string) Node[T] {
	h.mu.RLock()
	defer h.mu.RUnlock()

	k := h.hash(key)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] > k })
	if index >= len(h.nodes) {
		index = 0
	}
	return h.nodes[index]
}
