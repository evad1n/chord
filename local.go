package main

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Local unexported node functions

// Find returns the address of the node responsible for the given key
func (a NodeActor) find(
	req struct {
		key   Key
		start Address
	},
	addr *Address,
) error {
	var err error
	a.wait(func(n *Node) {
		hash := req.key.hashed()
		result := &AddressResult{false, n.Address}
		i := 0
		// Check locally
		// call(localAddress, "NodeActor.FindSuccessor", req.key, result)
		for !result.Found && i < maxRequests {
			if err := call(
				result.Address,
				"Node.FindSuccesor",
				hash,
				result,
			); err != nil {
				err = fmt.Errorf("find node: %v", err)
				return
			}
			i++
		}
		if !result.Found {
			err = errors.New("could not find node responsible for the key")
		}
	})
	return err
}

// Stabilize does something
func (a NodeActor) stabilize(request None, reply *None) error {
	return nil
}

// FixFingers makes the finger table correct
func (a NodeActor) FixFingers(request None, reply *None) error {
	return nil
}

// Stringer interface for Node dump
func (n Node) String() string {
	var w strings.Builder
	w.WriteString("DUMP: Node info\n\n")
	w.WriteString(fmt.Sprintf("Address: %s\n", n.Address))

	w.WriteString("\nData items:\n")
	for key, value := range n.Data {
		w.WriteString(fmt.Sprint(KeyValue{key, value}))
	}

	return w.String()
}

func (kv KeyValue) String() string {
	return fmt.Sprintf("%20s => %s\n", kv.Key, kv.Value)
}

// Returns true if elt is between start and end on the ring
func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}
