package main

import (
	"fmt"
	"math/big"
	"strings"
)

// Stringer interface for Node dump
func (n Node) String() string {
	var w strings.Builder
	w.WriteString("DUMP: Node info\n")
	w.WriteString(fmt.Sprintf("Address: %s\n", n.Address))

	w.WriteString("Data items:")
	for key, value := range n.Data {
		w.WriteString(fmt.Sprint(KeyValue{key, value}))
	}

	return w.String()
}

// Returns true if elt is between start and end on the ring
func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}
