package consistentHash

import (
	"testing"
)

// MockHasher is a mock implementation of the Hasher interface for testing.
type MockHasher struct{}

func (m *MockHasher) Hash(key string) int {
	// Simple mock hash function for testing purposes.
	// In a real-world scenario, you would use a proper hash function.
	return len(key)
}

func TestConsistentHash_AddNode(t *testing.T) {
	hasher := &MockHasher{}
	ch := New(hasher)

	node := Node{name: "node1", host: "127.0.0.1"}
	key := ch.AddNode(node)

	if key != hasher.Hash(node.host)%maxNodes {
		t.Errorf("Expected key to be %d, got %d", hasher.Hash(node.host)%maxNodes, key)
	}

	if len(ch.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(ch.nodes))
	}
}

func TestConsistentHash_PickNode(t *testing.T) {
	hasher := &MockHasher{}
	ch := New(hasher)

	node1 := Node{name: "node1", host: "127.0.0.1"}
	node2 := Node{name: "node2", host: "192.168.1.1"}

	ch.AddNode(node1)
	ch.AddNode(node2)

	pickedNode := ch.PickNode("key123")

	// Assuming a simple hash function that returns the length of the key,
	// and given the mock nodes added, the picked node should be node1.
	if pickedNode.name != "node1" {
		t.Errorf("Expected node1, got %s", pickedNode.name)
	}
}

func TestConsistentHash_hash(t *testing.T) {
	hasher := &MockHasher{}
	ch := New(hasher)

	key := "key123"
	expectedHash := hasher.Hash(key) % maxNodes
	actualHash := ch.hash(key)

	if actualHash != expectedHash {
		t.Errorf("Expected hash to be %d, got %d", expectedHash, actualHash)
	}
}
