package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"time"
)

const (
	maxSuccessors    = 5
	numFingerEntries = 161

	stabilizeInterval        = 2
	fixFingersInterval       = 5
	checkPredecessorInterval = 5
)

var next = 0 // The next entry in the finger table to fix

// Run stabilize, fix fingers, and check predecessor in background goroutines
func (n *Node) startBackgroundMaintenance() {
	// Stabilize
	log.Printf("Stabilizing every %d seconds", stabilizeInterval)
	go func() {
		if err := n.stabilize(); err != nil {
			log.Fatalf("initial stabilize: %v", err)
		}
		for range time.Tick(time.Second * stabilizeInterval) {
			if err := n.stabilize(); err != nil {
				log.Fatalf("stabilize: %v", err)
			}
		}
	}()
	// FixFingers
	log.Printf("Fixing fingers every %d seconds", fixFingersInterval)
	go func() {
		if err := n.fixFingers(); err != nil {
			log.Fatalf("initial fix fingers: %v", err)
		}
		for range time.Tick(time.Second * fixFingersInterval) {
			if err := n.fixFingers(); err != nil {
				log.Fatalf("fix fingers: %v", err)
			}
		}
	}()
	// CheckPredecessor
	log.Printf("Checking predecessor every %d seconds", checkPredecessorInterval)
	go func() {
		if err := n.checkPredecessor(); err != nil {
			log.Fatalf("initial check predecessor: %v", err)
		}
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
	if err := call(*n.Successors[0], "NodeActor.GetNodeLinks", None{}, &links); err != nil {
		// Cut off that one from list
		n.Successors = n.Successors[1:]
		if len(n.Successors) == 0 {
			// No successors so set successor to ourself
			n.Successors = append(n.Successors, &n.Address)
		}
	} else {
		// Update successor links

		// Add current node's first successor to successor's successor list
		// Prepend current successor to a slice of the successors
		n.Successors = append([]*Address{n.Successors[0]}, links.Successors...)
		// Truncate if necessary
		if len(n.Successors) > maxSuccessors {
			n.Successors = n.Successors[:maxSuccessors]
		}

		// Update predecessor links

		// FIX: what to do with single node ring predecessor (is nil)
		// Check if our successor's predecessor isn't us
		if n.Predecessor != nil && between(n.Hash, links.Predecessor.hashed(), n.Successors[0].hashed(), false) {
			// Set our successor to be this node in between now
			n.Successors[0] = links.Predecessor
		}
	}
	// Notify successor to check its predecessor
	if err := call(*n.Successors[0], "NodeActor.Notify", n.Address, &None{}); err != nil {
		return fmt.Errorf("notifying successor: %v", err)
	}

	return nil
}

/* // called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
n.fix fingers()
next = next + 1 ;
if (next > m)
next = 1 ;
finger[next] = find successor(n + 2
next−1
); */
func (n *Node) fixFingers() error {
	next++
	if next >= numFingerEntries {
		next = 1
	}
	address, err := find(n.jump(next), &n.Address)
	if err != nil {
		return fmt.Errorf("finding finger table entry: %v", err)
	}
	// Optimization because sparse nodes mean the successor for each entry is probably the same
	for next < numFingerEntries && between(n.Hash, n.jump(next), address.hashed(), false) {
		n.Fingers[next] = address
		next++
	}
	return nil
}

// Verify predecessor is still functional
func (n *Node) checkPredecessor() error {
	if n.Predecessor == nil {
		log.Println("no predecessor")
		return nil
	}
	var success bool
	if err := call(*n.Predecessor, "NodeActor.Ping", None{}, &success); err != nil || !success {
		log.Printf("failed to contact predecessor: %v\n", err)
		n.Predecessor = nil
	}
	return nil
}

// IDK
const keySize = sha1.Size * 8

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

// This computes the hash of a position across the ring that should be pointed to by the given finger table entry (using 1-based numbering).
func (n *Node) jump(fingerentry int) *big.Int {
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n.Hash, jump)

	return new(big.Int).Mod(sum, hashMod)
}
