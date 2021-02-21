package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FIX: should local node be stored or called from RPC?

var (
	localHost    string
	localPort    = 3410  // Port to listen on
	localAddress Address // The full address which is set upon joining
	joined       = false // Whether this node is part of a ring yet
)

func main() {
	localHost = getLocalAddress()
	fmt.Printf("Current address: %s\n", localHost)
	fmt.Printf("Current port: %s\n", localPort)

	createMaps()
	defaultCommands()

	commandLoop()
}

func commandLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print(">>> ")
	for scanner.Scan() {
		// Otherwise process commands
		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if cmd, exists := commands[strings.ToLower(words[0])]; exists {
				params := strings.Join(words[1:], " ")
				if err := cmd.do(params); err != nil {
					fmt.Println(ansiWrap(err.Error(), ansiColors["red"]))
				}
			} else {
				fmt.Println(ansiWrap("Unrecognized command!", ansiColors["red"]))
			}
		}
		// New prompt
		fmt.Print(">>> ")
	}
}
