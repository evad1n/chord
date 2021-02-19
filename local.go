package main

import "math/big"

func (n Node) address() string {
	return n.Host + ":" + n.Port
}

func (n Node) toString() string {
	var w strings.Builder
	w.WriteString("DUMP: Node info\n")
	w.WriteString(fmt.Sprintf("Full Address: %s\n", n.address()))
	w.WriteString(fmt.Sprintf("Host: %s\n", n.Host))
	w.WriteString(fmt.Sprintf("Port: %s\n", n.Port))

	w.WriteString("Data items:")
	for key, value := range currentNode.Data {
		w.WriteString(fmt.Sprintf("%s => %s", key, value))
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

func findSuccesor(id) (string, error) {
	// ask node n to find the successor of id
    // or a better node to continue the search with
    n.find_successor(id)
	if (id ∈ (n, successor])
	return true, successor;
	else
	return false, closest_preceding_node(id);
	
    // search the local table for the highest predecessor of id
    n.closest_preceding_node(id)
	// skip this loop if you do not have finger tables implemented yet
	for i = m downto 1
	if (finger[i] ∈ (n,id])
	return finger[i];
	return successor;
	
    // find the successor of id
    find(id, start)
	found, nextNode = false, start;
	i = 0
	while not found and i < maxSteps
	found, nextNode = nextNode.find_successor(id);
	i += 1
	if found
	return nextNode;
	else
	report error;
}