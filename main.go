package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	host string
	port = "3410" // Port to listen on

	currentNode *Node // The node running in this instance

	joined = false // Whether this node is part of a ring yet
)

func main() {
	host = getLocalAddress()
	fmt.Printf("Current address: %s\n", host)
	fmt.Printf("Current port: %s\n", port)

	createMaps()
	defaultCommands()

	commandLoop()
}

func commandLoop() {
	scanner := bufio.NewScanner(os.Stdout)

	fmt.Print(">>> ")
	for scanner.Scan() {
		// Otherwise process commands
		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if cmd, exists := commands[strings.ToLower(words[0])]; exists {
				params := strings.Join(words[1:], " ")
				cmd.do(params)
			} else {
				fmt.Println(ansiWrap("Unrecognized command!", ansiColors["red"]))
			}
		}
		// New prompt
		fmt.Print(">>> ")
	}
}
