package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	localHost string
	localPort = 3410  // Port to listen on
	localNode *Node   // The local node, only set after join/creation
	joined    = false // Whether this node is part of a ring yet
)

func main() {
	// DEBUGGING
	log.SetFlags(log.Lshortfile)

	fmt.Print("Welcome to the CHORD distributed hash table(DHT)\n\n")

	localHost = getLocalAddress()
	fmt.Printf("Current address: %s\n", localHost)
	fmt.Printf("Current port: %d\n", localPort)

	fmt.Println()

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
				switch {
				case !joined && cmd.joinRequired:
					fmt.Println(ansiWrap("must join a ring for this command", ansiColors["red"]))
					break
				default:
					if err := cmd.do(params); err != nil {
						fmt.Println(ansiWrap(err.Error(), ansiColors["red"]))
					}
				}
			} else {
				fmt.Println(ansiWrap("Unrecognized command!", ansiColors["red"]))
			}
		}
		// New prompt
		fmt.Print(">>> ")
	}
}
