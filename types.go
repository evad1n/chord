package main

import (
	"fmt"
	"math/big"
)

type (
	// Node is a part of the chord ring
	Node struct {
		Address     Address
		Hash        *big.Int
		Successors  [numSuccessors]Address
		Predecessor Address
		Fingers     [160]Address   // The finger table pointing to other nodes on the ring
		Data        map[Key]string // The data items stored at this node
	}

	// Hashable can be hashed and implements fmt.Stringer
	Hashable interface {
		hashed() *big.Int
		fmt.Stringer
	}

	// These will implement Hashable

	// Address represents an IPv4 address and a port following the form <Ipv4>:<port>
	Address string
	// Key represents a map key, which will be hashed
	Key string
	// Hash is just a wrapper for a *big.Int
	Hash struct {
		*big.Int
	}

	// KeyValue is a data item to be stored
	KeyValue struct {
		Key   Key
		Value string
	}

	// AddressResult represents a return address and if that address is the desired address
	AddressResult struct {
		Found   bool // Whether the returned address is a final or intermediate step
		Address Address
	}

	// None is a null value
	None struct{}
)
