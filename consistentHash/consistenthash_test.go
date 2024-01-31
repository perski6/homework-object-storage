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

// MockClient is a mock type that we will use for the generic type T in our tests.
type MockClient struct {
	ID string
}

func TestConsistentHash_AddNode(t *testing.T) {
	hasher := &MockHasher{}
	ch := New[MockClient](hasher)

	node := Node[MockClient]{Name: "node1", Client: MockClient{ID: "client1"}}
	key := ch.AddNode(node)

	if key != hasher.Hash(node.Name)%maxNodes {
		t.Errorf("Expected key to be %d, got %d", hasher.Hash(node.Name)%maxNodes, key)
	}

	if len(ch.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(ch.nodes))
	}
}

func TestConsistentHash_RemoveNode(t *testing.T) {
	hasher := &MockHasher{}
	ch := New[MockClient](hasher)

	node := Node[MockClient]{Name: "node1", Client: MockClient{ID: "client1"}}
	ch.AddNode(node)
	ch.RemoveNode(node.Name)

	if len(ch.nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(ch.nodes))
	}
}

func TestConsistentHash_PickNode(t *testing.T) {
	hasher := &MockHasher{}
	ch := New[MockClient](hasher)

	node1 := Node[MockClient]{Name: "node1", Client: MockClient{ID: "client1"}}
	node2 := Node[MockClient]{Name: "node2", Client: MockClient{ID: "client2"}}

	ch.AddNode(node1)
	ch.AddNode(node2)

	pickedNode := ch.PickNode("key123")

	// Assuming a simple hash function that returns the length of the key,
	// and given the mock nodes added, the picked node should be node1.
	if pickedNode.Name != "node1" {
		t.Errorf("Expected node1, got %s", pickedNode.Name)
	}
}

func TestConsistentHash_hash(t *testing.T) {
	hasher := &MockHasher{}
	ch := New[MockClient](hasher)

	key := "key123"
	expectedHash := hasher.Hash(key) % maxNodes
	actualHash := ch.hash(key)

	if actualHash != expectedHash {
		t.Errorf("Expected hash to be %d, got %d", expectedHash, actualHash)
	}
}
