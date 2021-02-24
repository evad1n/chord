package main

import (
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
)

const (
	maxRequests = 32 // Maximum number of requests a single lookup can generate
)

// Start the RPC server on the node
func (n *Node) startNode() error {
	// Make sure port isn't in use frst
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(localPort))
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}
	actor := n.startActor()
	rpc.Register(actor)
	rpc.HandleHTTP()
	go http.Serve(listener, nil)
	return nil
}

func (n *Node) startActor() NodeActor {
	ch := make(chan handler)
	// Launch actor channel
	go func() {
		for evt := range ch {
			evt(n)
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
		return err
	}
	defer client.Close()

	// Synchronous call
	if err := client.Call(method, request, reply); err != nil {
		return err
	}

	return nil
}

// Ping simply tests an RPC connection
func (a NodeActor) Ping(_ None, reply *bool) error {
	a.wait(func(n *Node) {
		*reply = true
	})
	return nil
}

// FindSuccessor asks the node to find the successor of an id, or a better node to continue the search with
func (a NodeActor) FindSuccessor(id *big.Int, result *AddressResult) error {
	a.wait(func(n *Node) {
		// If it is between us and our successor
		if between(n.Hash, id, n.Successors[0].hashed(), true) {
			*result = AddressResult{
				Found:   true,
				Address: n.Successors[0],
			}
		} else {
			*result = AddressResult{
				Found:   false,
				Address: n.closestPrecedingNode(id),
			}
		}
	})
	return nil
}

// Notify signals a node that another node thinks it should be its predecessor
func (a NodeActor) Notify(address Address, _ *None) error {
	a.wait(func(n *Node) {
		if n.Predecessor == nil || between(n.Predecessor.hashed(), address.hashed(), n.Hash, false) {
			n.Predecessor = &address
		}
	})
	return nil
}

// GetNodeLinks returns the successors and predecessor of a node
func (a NodeActor) GetNodeLinks(request None, links *NodeLink) error {
	a.wait(func(n *Node) {
		*links = NodeLink{
			Predecessor: n.Predecessor,
			Successors:  n.Successors,
		}
	})
	return nil
}

// Put adds an item to the database
func (a NodeActor) Put(kv KeyValue, _ *None) error {
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

// PutAll adds all key/value pairs in a map to the local data
func (a NodeActor) PutAll(data map[Key]string, _ *None) error {
	var err error
	a.wait(func(n *Node) {
		for key, value := range data {
			n.Data[key] = value
		}
	})
	return err
}

// GetAll gathers all key/value pairs from a node and transfers them to a newly joined node that they belong to
func (a NodeActor) GetAll(newAddress Address, data *map[Key]string) error {
	var err error
	transfer := make(map[Key]string)
	a.wait(func(n *Node) {
		for key, value := range n.Data {
			if between(newAddress.hashed(), key.hashed(), n.Successors[0].hashed(), true) {
				transfer[key] = value
				delete(n.Data, key)
			}
		}
	})
	return err
}

// Dump delivers all info on a node
func (a NodeActor) Dump(_ None, dumpReturn *DumpReturn) error {
	a.wait(func(n *Node) {
		*dumpReturn = DumpReturn{
			Dump:      n.String(),
			Successor: n.Successors[0],
		}
	})
	return nil
}
