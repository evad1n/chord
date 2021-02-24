package main

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Local unexported node functions

// Find returns the address of the node responsible for the given id.
// Node agnostic, just acts on a ring
func find(id *big.Int, start *Address) (*Address, error) {
	result := AddressResult{
		Found:   false,
		Address: start,
	}
	i := 0
	for !result.Found && i < maxRequests {
		if err := call(*result.Address, "NodeActor.FindSuccessor", id, &result); err != nil {
			return result.Address, fmt.Errorf("find successor: %v", err)
		}
		i++
	}
	if result.Found {
		return result.Address, nil
	}
	return result.Address, errors.New("exceeded max lookups")
}

// Search local fingers for highest predecessor of id
func (n *Node) closestPrecedingNode(id *big.Int) *Address {
	for i := numFingerEntries; i > 0; i-- {
		if n.Fingers[i] == nil {
			continue
		}
		if between(n.Hash, n.Fingers[i].hashed(), id, false) {
			return n.Fingers[i]
		}
	}
	// Otherwise just return successor
	return n.Successors[0]
}

// Create local node instance
func createNode() *Node {
	return &Node{
		Address:     Address(localHost + ":" + fmt.Sprint(localPort)),
		Hash:        Address(localHost + ":" + fmt.Sprint(localPort)).hashed(),
		Predecessor: nil,
		Data:        make(map[Key]string),
	}
}

// Create a new chord ring
func createRing() (*Node, error) {
	n := createNode()
	if err := n.startNode(); err != nil {
		return n, fmt.Errorf("starting node RPC server: %v", err)
	}
	// Set successor to itself
	n.Successors = append(n.Successors, &n.Address)
	// Start background tasks
	n.startBackgroundMaintenance()
	return n, nil
}

// Join an existing chord ring
func joinRing(joinAddress Address) (*Node, error) {
	localAddress := Address(localHost + ":" + fmt.Sprint(localPort))
	n := createNode()
	// Call find starting at supplied address, searching for local address
	successor, err := find(localAddress.hashed(), &joinAddress)
	if err != nil {
		return nil, fmt.Errorf("finding place on ring: %v", err)
	}
	n.Successors = append(n.Successors, successor)
	// Now start server
	if err := n.startNode(); err != nil {
		return n, fmt.Errorf("starting node RPC server: %v", err)
	}
	// Start background tasks
	n.startBackgroundMaintenance()
	// Request

	return n, nil
}

// Returns true if elt is between start and end on the ring
func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

// Stringer interface for Node dump
func (n Node) String() string {
	var w strings.Builder
	w.WriteString("DUMP: Node info\n\n")
	w.WriteString(fmt.Sprintf("Predecessor: %s\n\n", n.Predecessor))
	w.WriteString(fmt.Sprintf("Address: %s\n\n", n.Address))
	for i, successor := range n.Successors {
		w.WriteString(fmt.Sprintf("Sucessor[%d]: %s\n", i, successor))
	}

	// Print only unique finger table entries
	unique := make(map[*Address]int)
	for i, address := range n.Fingers {
		unique[address] = i
	}
	w.WriteString("\nFinger table:\n")
	for address, i := range unique {
		if address != nil {
			w.WriteString(fmt.Sprintf("   [%d]: %s\n", i, address))
		}
	}

	if len(n.Data) > 0 {
		w.WriteString("\nData items:\n")
		for key, value := range n.Data {
			w.WriteString(fmt.Sprintf("   %s\n", KeyValue{key, value}))
		}
	} else {
		w.WriteString("\nNo data items")
	}

	w.WriteString("\n")

	return w.String()
}

// Print a key value pair
func (kv KeyValue) String() string {
	return fmt.Sprintf("%20s => %s", kv.Key, kv.Value)
}
