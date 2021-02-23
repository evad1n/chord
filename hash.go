package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

func (a Address) hashed() *big.Int {
	return hashString(string(a))
}

func (a Address) String() string {
	return fmt.Sprintf("%s [ %s ]", readableHash(a.hashed()), string(a))
}

func (k Key) hashed() *big.Int {
	return hashString(string(k))
}

func (k Key) String() string {
	return fmt.Sprintf("%s [ %s ]", readableHash(k.hashed()), string(k))
}

func readableHash(hash *big.Int) string {
	return fmt.Sprintf("%040x", hash)[:8] + "..."
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}
