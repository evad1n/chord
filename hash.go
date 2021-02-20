package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

func (a Address) hashed() *big.Int {
	return hashString(string(a))
}

func (k Key) hashed() *big.Int {
	return hashString(string(k))
}

func (h Hash) String() string {
	return fmt.Sprintf("%040x", h)[:8] + "..."
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func (kv KeyValue) String() string {
	return fmt.Sprintf("%8s => %s", kv.Key, kv.Value)
}
