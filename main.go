package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	localHost string
	localPort = 3400  // Port to listen on
	localNode *Node   // The local node, only set after join/creation
	joined    = false // Whether this node is part of a ring yet

	logging = false // Whether to print log messages
)

// A way to color the log yellow
type myWriter struct {
	w io.Writer
}

func (w myWriter) Write(p []byte) (n int, err error) {
	if !logging {
		return
	}
	w.w.Write([]byte(ansiColors["yellow"]))
	n, err = w.w.Write(p)
	w.w.Write([]byte("\x1b[0m"))
	return
}

func main() {
	// Setup
	rand.Seed(time.Now().Unix())
	createMaps()
	defaultCommands()

	// DEBUGGING
	log.SetFlags(log.Lshortfile)
	log.SetOutput(myWriter{os.Stdout})

	fmt.Print("Welcome to the CHORD distributed hash table(DHT)\n\n")

	localHost = getLocalAddress()
	fmt.Printf("Current address: %s\n", localHost)
	fmt.Printf("Current port: %d\n", localPort)
	if logging {
		fmt.Println("Logging is turned ON")
	} else {
		fmt.Println("Logging is turned OFF")
	}
	fmt.Println()

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
				default:
					if err := cmd.do(params); err != nil {
						fmt.Println(ansiWrap(err.Error(), ansiColors["red"]))
					} else {
						fmt.Println(ansiWrap("OK", ansiColors["green"]))
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
