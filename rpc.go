package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

const (
	numSuccessors = 5
	maxRequests   = 32 // Maximum number of requests a single lookup can generate
)

// Start the RPC server
func startNode() {
	currentNode = &Node{
		Address: Address(host + ":" + port),
		Hash:    currentNode.Address.hashed(),
		Data:    make(map[Key]string),
	}
	rpc.Register(currentNode)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+port)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	// Run stabilize, fix fingers, and check predecessor in goroutines
	go http.Serve(l, nil)
}

// The RPC call
func call(address Address, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", string(address))
	if err != nil {
		log.Printf("rpc call dialing: %v", err)
		return err
	}
	defer client.Close()

	// Synchronous call
	if err := client.Call(method, request, reply); err != nil {
		log.Printf("client rpc call '%s': %v", method, err)
		return err
	}

	return nil
}

// Ping simply tests an RPC connection
func (n *Node) Ping(request None, reply *bool) error {
	*reply = true
	return nil
}

// Find returns the address of the node responsible for the given key
func (n *Node) find(key Key) (Address, error) {
	hash := key.hashed()
	result := &AddressResult{false, n.Address}
	i := 0
	// Check locally
	currentNode.FindSuccesor(key, result)
	for !result.Found && i < maxRequests {
		if err := call(
			result.Address,
			"Node.FindSuccesor",
			hash,
			result,
		); err != nil {
			return n.Address, fmt.Errorf("find node: %v", err)
		}
		i++
	}
	if result.Found {
		return result.Address, nil
	}
	return n.Address, errors.New("could not find node responsible for the key")
}

// FindSuccesor finds the nearest successor node of they key with given id
func (n *Node) FindSuccesor(key Key, result *AddressResult) error {
	id := key.hashed()
	// If it is one of our successors
	if between(n.Hash, id, n.Successors[len(n.Successors)-1].hashed(), true) {
		// Loop from nearest to farthest to find successor
		for _, s := range n.Successors {
			// Triggers on the nearest successor
			if id.Cmp(s.hashed()) < 0 {
				*result = AddressResult{true, s}
			}
		}
	} else {
		// call(n.Successors[0], "Node.FindSuccessor", id, address)
		// Give address of last successor
		*result = AddressResult{false, n.Successors[len(n.Successors)-1]}
	}
	return nil
}

// Create creates a new chord ring with only this node in it
func (n *Node) create() {

}

// Join joins an existing chord ring containing the node at the address specified
func (n *Node) Join(address string, reply *None) error {
	return nil
}

// Stabilize does something
func (n *Node) Stabilize(request None, reply *None) error {
	return nil
}

// Notify notifies the nodes around
func (n *Node) Notify(address string, reply *None) error {
	return nil
}

// FixFingers makes the finger table correct
func (n *Node) FixFingers(request None, reply *None) error {
	return nil
}

// CheckPredecessor checks to see if the predecessor is correct
func (n *Node) CheckPredecessor(request None, reply *None) error {
	return nil
}

// Put adds an item to the database
func (n *Node) Put(data *KeyValue, reply *None) error {
	n.Data[data.Key] = data.Value
	return nil
}

// Get retrieves the value of a key in the database
func (n *Node) Get(key Key, value *string) error {
	*value = n.Data[key]
	return nil
}

// Delete removes a key and its associated value from the database
func (n *Node) Delete(key Key, value *string) error {
	*value = n.Data[key]
	delete(n.Data, key)
	return nil
}
