package main

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/perski6/homework-object-storage/consistentHash"
)

func main() {
	ch := consistentHash.New(sha256Hasher{})
}

type sha256Hasher struct {
}

func (h sha256Hasher) Hash(key string) int {
	hash := sha256.New()
	hash.Write([]byte(key))
	return hex.EncodeToString(hash.Sum(nil))
}
