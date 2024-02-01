package consistentHash

import (
	"sort"
	"sync"
)

type Status string

const (
	Running Status = "Running"
	Stopped Status = "Stopped"
)
const maxNodes int = 1337

// NodeProvider is an interface that defines methods to interact with a consistent hash ring.
type NodeProvider[T any] interface {
	PickNode(key string) Node[T]
	AddNode(n Node[T]) int
	RemoveNode(host string)
	StopNode(host string)
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
	Status Status
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

func (h *ConsistentHash[T]) AddNode(n Node[T]) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.hash(n.Name)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] >= key })
	if index < len(h.keys) && h.keys[index] == key {
		// Key already exists, do not add again just set the status to Running
		h.nodes[index].Status = Running
		return key
	}
	// Insert node at the correct position
	h.nodes = insertNode(h.nodes, index, n)
	// Insert key at the correct position
	h.keys = insertKey(h.keys, index, key)

	return key
}

func (h *ConsistentHash[T]) StopNode(host string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.hash(host)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] >= key })
	h.nodes[index].Status = Stopped
}
func insertNode[T any](nodes []Node[T], index int, node Node[T]) []Node[T] {
	nodes = append(nodes, Node[T]{})     // Make space for the new node
	copy(nodes[index+1:], nodes[index:]) // Shift nodes to the right
	nodes[index] = node                  // Insert new node
	return nodes
}

func insertKey(keys []int, index int, key int) []int {
	keys = append(keys, 0)             // Make space for the new key
	copy(keys[index+1:], keys[index:]) // Shift keys to the right
	keys[index] = key                  // Insert new key
	return keys
}

func (h *ConsistentHash[T]) RemoveNode(host string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.hash(host)
	index := sort.Search(len(h.keys), func(i int) bool { return h.keys[i] >= key })
	if index < len(h.keys) && h.keys[index] == key {
		// Remove node and key at the correct position
		h.nodes = removeNode(h.nodes, index)
		h.keys = removeKey(h.keys, index)
	}
}

func removeNode[T any](nodes []Node[T], index int) []Node[T] {
	return append(nodes[:index], nodes[index+1:]...)
}

func removeKey(keys []int, index int) []int {
	return append(keys[:index], keys[index+1:]...)
}
