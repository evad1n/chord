package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type (
	command struct {
		name        string
		description string
		usage       string
		do          func(string)
	}
)

var (
	commands   map[string]command // Map of command aliases to commands
	ansiColors map[string]string  // ANSI colors to code map
)

// Initialize and populate lookup tables
func createMaps() {

	// ANSI colors map
	ansiColors = make(map[string]string)
	ansiColors["black"] = "\x1b[30m"
	ansiColors["red"] = "\x1b[31m"
	ansiColors["green"] = "\x1b[32m"
	ansiColors["yellow"] = "\x1b[33m"
	ansiColors["blue"] = "\x1b[34m"
	ansiColors["magenta"] = "\x1b[35m"
	ansiColors["cyan"] = "\x1b[36m"
	ansiColors["white"] = "\x1b[37m"
}

/* Maps prefixes to full name for a map */
func addMapPrefix(full string, m map[string]string) {
	for i := range full {
		if i == 0 {
			continue
		}
		prefix := full[:i]
		if _, exists := m[prefix]; !exists {
			m[prefix] = full
		}
	}
	m[full] = full
}

// Adds all starting commands
func defaultCommands() {
	commands = make(map[string]command)

	commands["help"] = command{
		name:        "help",
		description: "List all commands and descriptions",
		do:          listCommands,
	}
	commands["quit"] = command{
		name:        "quit",
		description: "Quit and offload node data gracefully",
		do:          quit,
	}
	commands["port"] = command{
		name:        "port",
		description: "Change the listening port",
		usage:       "port <number>",
		do:          changePort,
	}
	commands["ping"] = command{
		name:        "ping",
		description: "Ping another node",
		usage:       "ping <host>:<port>",
		do:          ping,
	}
	commands["create"] = command{
		name:        "create",
		description: "Create and join a new chord ring",
		do:          create,
	}
	commands["join"] = command{
		name:        "join",
		description: "Join a chord ring from a known node address",
		usage:       "join <host>:<port>",
		do:          join,
	}
	commands["put"] = command{
		name:        "put",
		description: "Add a key/value pair to the database",
		usage:       "put <key> <value>",
		do:          put,
	}
	commands["get"] = command{
		name:        "get",
		description: "Get the value of a key",
		usage:       "get <key>",
		do:          get,
	}
	commands["delete"] = command{
		name:        "delete",
		description: "Dlete a key and its associated value",
		usage:       "delete <key>",
		do:          deleteKey,
	}
	commands["putrandom"] = command{
		name:        "putrandom",
		description: "Add random data items to the database",
		usage:       "putrandom <num_items>",
		do:          putRandom,
	}
	// Information/debugging
	commands["dump"] = command{
		name:        "dump",
		description: "Dumps current node information",
		do:          dump,
	}
	commands["dumpkey"] = command{
		name:        "dumpkey",
		description: "Dumps info on the node responsible for a key",
		usage:       "dumpkey <key>",
		do:          dumpKey,
	}
	commands["dumpaddr"] = command{
		name:        "dumpaddr",
		description: "Dumps info on the node at the requested address",
		usage:       "dumpaddr <host>:<port>",
		do:          dumpAddress,
	}
	commands["dumpall"] = command{
		name:        "dumpall",
		description: "Dumps info on each node in the current ring",
		do:          dumpAll,
	}
}

//////////////
// Commands //
//////////////

// Lists known aliases for commands
func listCommands(_ string) {
	var w strings.Builder

	w.WriteString(fmt.Sprintf("+%s+\n", strings.Repeat("-", 98)))
	w.WriteString(fmt.Sprintf("|%s|\n", centerText("COMMANDS LIST", 98, ' ')))
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", 12),
		strings.Repeat("-", 52),
		strings.Repeat("-", 32),
	))
	w.WriteString(fmt.Sprintf(
		"| %-10s | %-50s | %-20s |\n",
		centerText("NAME", 10, ' '),
		centerText("DESCRIPTION", 50, ' '),
		centerText("USAGE", 30, ' '),
	))
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", 12),
		strings.Repeat("-", 52),
		strings.Repeat("-", 32),
	))

	cmds := []command{}
	for _, v := range commands {
		cmds = append(cmds, v)
	}

	// Sort by name
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].name < cmds[j].name
	})

	for _, c := range cmds {
		w.WriteString(fmt.Sprintf("| %-10s | %-50s |", c.name, c.description))
		if c.usage != "" {
			w.WriteString(fmt.Sprintf(" %-30s |\n", c.usage))
		} else {
			w.WriteString(fmt.Sprintf(" %-30s |\n", c.name))
		}
	}
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", 12),
		strings.Repeat("-", 52),
		strings.Repeat("-", 32),
	))

	fmt.Println(w.String())
}

// Quit gracefully and offload data to other nodes
func quit(_ string) {
	fmt.Println("Quitting...")
	os.Exit(0)
}

// Change port to listen on, can't be done after joining
func changePort(p string) {
	if !joined {
		if newPort, err := strconv.Atoi(p); err == nil {
			fmt.Printf("Listening port changed from %s to %d\n", port, newPort)
			port = p
		} else {
			fmt.Println(ansiWrap("bad port", ansiColors["red"]))
		}
	} else {
		fmt.Println(ansiWrap("can't change port. already listening", ansiColors["red"]))
	}
}

func ping(address string) {
	if addr, err := validateAddress((address)); err == nil {
		fmt.Printf("Attempting to ping %s...\n", address)
		var success bool
		call(addr, "Node.Ping", None{}, &success)
		if success {
			fmt.Println("Success")
		} else {
			fmt.Println(ansiWrap("Error", ansiColors["red"]))
		}
	} else {
		fmt.Printf(ansiWrap(err.Error(), ansiColors["red"]))
	}
}

func create(_ string) {
	if !joined {
		go startNode()
		joined = true
	} else {
		fmt.Println(ansiWrap("can't create ring. already part of a ring", ansiColors["red"]))
	}
}

func join(address string) {
	if !joined {
		if addr, err := validateAddress(address); err == nil {
			go startNode()
			fmt.Println(addr, addr.hashed())
			joined = true
		} else {
			fmt.Println(ansiWrap(err.Error(), ansiColors["red"]))
		}
	} else {
		fmt.Println(ansiWrap("can't join ring. already part of a ring", ansiColors["red"]))
	}
}

func dump(_ string) {
	if currentNode != nil {
		fmt.Println(currentNode.toString())
	} else {
		fmt.Println("Node does not exist yet")
	}
}

func dumpKey(key string) {

}

func dumpAddress(address string) {

}

func dumpAll(_ string) {

}

func put(data string) {
	if words := strings.Fields(data); len(words) == 2 {
		key, value := Key(words[0]), words[1]
		fmt.Printf("%s => %s", key, value)
		kv := KeyValue{key, value}
		address, err := currentNode.find(key)
		if err != nil {
			log.Printf("put: %v", err)
		}

		if err := call(address, "Node.Put", kv, &None{}); err != nil {
			log.Printf("put: %v", err)
		}
		fmt.Println("successful put: ", kv)
	} else {
		fmt.Println("Too many values: <key> <value>")
	}
}

func get(input string) {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(input)
		fmt.Printf("Get item with key: %s", key)
		address, err := currentNode.find(key)
		if err != nil {
			log.Printf("get finding: %v", err)
		}
		var value *string
		if err := call(address, "Node.Get", key, value); err != nil {
			log.Printf("get getting: %v", err)
		}
		fmt.Println(KeyValue{key, *value})
	} else {
		fmt.Println("Too many values: <key>")
	}
}

func deleteKey(input string) {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(input)
		fmt.Printf("Delete item with key: %s", key)
		address, err := currentNode.find(key)
		if err != nil {
			log.Printf("delete finding: %v", err)
		}
		var value *string
		if err := call(address, "Node.Delete", key, value); err != nil {
			log.Printf("delete deleting: %v", err)
		}
		fmt.Printf("Successfully deleted item with key: %s, and value %s", key, *value)
	} else {
		fmt.Println("Too many values: <key>")
	}
}

func putRandom(count string) {

}
