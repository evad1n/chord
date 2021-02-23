package main

import (
	"fmt"
	"log"
	"time"
)

const (
	maxSuccessors = 5

	stabilizeInterval        = 2
	fixFingersInterval       = 5
	checkPredecessorInterval = 5
)

// Run stabilize, fix fingers, and check predecessor in background goroutines
func (n *Node) startBackgroundMaintenance() {
	// Stabilize
	log.Printf("Stabilizing every %d seconds", stabilizeInterval)
	go func() {
		for range time.Tick(time.Second * stabilizeInterval) {
			if err := n.stabilize(); err != nil {
				log.Fatalf("stabilize: %v", err)
			}
		}
	}()
	// FixFingers
	log.Printf("Fixing fingers every %d seconds", fixFingersInterval)
	go func() {
		for range time.Tick(time.Second * fixFingersInterval) {
			if err := n.fixFingers(); err != nil {
				log.Fatalf("fix fingers: %v", err)
			}
		}
	}()
	// CheckPredecessor
	log.Printf("Checking predecessor every %d seconds", checkPredecessorInterval)
	go func() {
		for range time.Tick(time.Second * checkPredecessorInterval) {
			if err := n.checkPredecessor(); err != nil {
				log.Fatalf("check predecessor: %v", err)
			}
		}
	}()
}

// Maintain successor list correctly
func (n *Node) stabilize() error {
	var links NodeLink
	if err := call(n.Successors[0], "NodeActor.GetNodeLinks", None{}, &links); err != nil {
		// Can't contact successor => assume it failed
		// Cut off that one from list
		n.Successors = n.Successors[1:]
		if len(n.Successors) == 0 {
			// No successors so set successor to ourself
			n.Successors = append(n.Successors, n.Address)
		}
	} else {
		// Update successor links

		// Add current node's first successor to successor's successor list
		// Prepend current successor to a slice of the successors
		n.Successors = append([]Address{n.Successors[0]}, links.Successors...)
		// Truncate if necessary
		if len(n.Successors) > maxSuccessors {
			n.Successors = n.Successors[:maxSuccessors]
		}

		// Update predecessor links

		// FIX: what to do with single node ring predecessor (is nil)
		// Check if our successor's predecessor isn't us
		if n.Predecessor != nil && between(n.Hash, links.Predecessor.hashed(), n.Successors[0].hashed(), false) {
			// Set our successor to be this node in between now
			n.Successors[0] = *links.Predecessor
		}
	}
	// Notify successor to check its predecessor
	if err := call(n.Successors[0], "NodeActor.Notify", n.Address, &None{}); err != nil {
		return fmt.Errorf("notifying successor: %v", err)
	}

	return nil
}

// Fix the finger tables
func (n *Node) fixFingers() error {
	return nil
}

// Verify correct successors and predecessor of node
func (n *Node) checkPredecessor() error {
	return nil
}
