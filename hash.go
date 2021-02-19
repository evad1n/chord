package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

type (
	Hashable interface {
		Hashed() *big.Int
	}
)

func (a Address) Hashed() *big.Int {
	return hashString(string(a))
}

func (k Key) Hashed() *big.Int {
	return hashString(string(k))
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func (kv KeyValue) String() string {
	hex := fmt.Sprintf("%040x", kv.Key)
	return fmt.Sprintf("%s.. (%s)", hex[:8], string(kv.Value))
}
