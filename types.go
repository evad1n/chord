package main

import (
	"fmt"
	"math/big"
)

type (
	// NodeActor represents an RPC actor for the Node client
	NodeActor chan<- handler
	// Some operation on a Node
	handler func(*Node)

	// Node is a part of the chord ring
	Node struct {
		Address     Address  // The string representation of an address (HOST:PORT)
		Hash        *big.Int // The hash of the address
		Successors  []Address
		Predecessor Address
		Fingers     [numFingerEntries]Address // The finger table pointing to addresses farther down the ring (increasing by powers of 2)
		Data        map[Key]string            // The data items stored at this node
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

	// RPC request/reply structs

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

	// NodeLink contains the predecessor and successor links for a node
	NodeLink struct {
		Predecessor Address
		Successors  []Address
	}

	// DumpReturn contains the dump info and the successor address
	DumpReturn struct {
		Dump      string // The string containing all the dump info
		Successor Address
	}

	// None is a null value
	None struct{}
)
