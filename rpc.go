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
func startNode() error {
	actor := startActor()
	rpc.Register(actor)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(localPort))
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}
	// TODO:
	// Run stabilize, fix fingers, and check predecessor in goroutines
	go http.Serve(listener, nil)
	return nil
}

func startActor() NodeActor {
	ch := make(chan handler)
	localNode = &Node{
		Address: Address(localHost + ":" + fmt.Sprint(localPort)),
		Hash:    Address(localHost + ":" + fmt.Sprint(localPort)).hashed(),
		Data:    make(map[Key]string),
	}
	// Launch actor channel
	go func() {
		for evt := range ch {
			evt(localNode)
		}
	}()
	return ch
}

// Blocks until actor executes
func (a NodeActor) wait(f handler) {
	done := make(chan None)
	a <- func(n *Node) {
		f(n)
		done <- None{}
	}
	<-done
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
func (a NodeActor) Ping(request None, reply *bool) error {
	a.wait(func(n *Node) {
		*reply = true
	})
	return nil
}

// FindSuccesor finds the nearest successor node of they key with given id
func (a NodeActor) FindSuccesor(key Key, result *AddressResult) error {
	a <- func(n *Node) {
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
			// TODO: loop through fingers and call FindSuccessor until found
			// call(n.Successors[0], "Node.FindSuccessor", id, address)
			// Give address of last successor
			*result = AddressResult{false, n.Successors[len(n.Successors)-1]}
		}
	}

	return nil
}

// Join joins an existing chord ring containing the node at the address specified
func (a NodeActor) Join(address string, reply *None) error {
	return nil
}

// Notify notifies the nodes around
func (a NodeActor) Notify(address string, reply *None) error {
	return nil
}

// CheckPredecessor checks to see if the predecessor is correct
func (a NodeActor) CheckPredecessor(request None, reply *None) error {
	return nil
}

// Put adds an item to the database
func (a NodeActor) Put(kv KeyValue, reply *None) error {
	a.wait(func(n *Node) {
		n.Data[kv.Key] = kv.Value
	})
	return nil
}

// Get retrieves the value of a key in the database
func (a NodeActor) Get(key Key, value *string) error {
	var err error
	a.wait(func(n *Node) {
		if val, exists := n.Data[key]; exists {
			*value = val
		} else {
			err = errors.New("no such key")
		}
	})
	return err
}

// Delete removes a key and its associated value from the database
func (a NodeActor) Delete(key Key, value *string) error {
	var err error
	a.wait(func(n *Node) {
		if val, exists := n.Data[key]; exists {
			*value = val
			delete(n.Data, key)
		} else {
			err = errors.New("no such key")
		}
	})
	return err
}

// Dump delivers all info on a node
func (a NodeActor) Dump(_ None, node *Node) error {
	a.wait(func(n *Node) {
		node = n
	})
	return nil
}
