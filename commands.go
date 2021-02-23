package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type (
	command struct {
		name         string
		description  string
		usage        string
		do           func(string) error
		joinRequired bool // Whether the command requires a ring to function
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
		name:         "put",
		description:  "Add a key/value pair to the database",
		usage:        "put <key> <value>",
		do:           put,
		joinRequired: true,
	}
	commands["get"] = command{
		name:         "get",
		description:  "Get the value of a key",
		usage:        "get <key>",
		do:           get,
		joinRequired: true,
	}
	commands["delete"] = command{
		name:         "delete",
		description:  "Delete a key and its associated value",
		usage:        "delete <key>",
		do:           deleteKey,
		joinRequired: true,
	}
	commands["putrandom"] = command{
		name:         "putrandom",
		description:  "Add random data items to the database",
		usage:        "putrandom <num_items>",
		do:           putRandom,
		joinRequired: true,
	}
	// Information/debugging
	commands["dump"] = command{
		name:         "dump",
		description:  "Dumps current node information",
		do:           dump,
		joinRequired: true,
	}
	commands["dumpkey"] = command{
		name:         "dumpkey",
		description:  "Dumps info on the node responsible for a key",
		usage:        "dumpkey <key>",
		do:           dumpKey,
		joinRequired: true,
	}
	commands["dumpaddr"] = command{
		name:        "dumpaddr",
		description: "Dumps info on the node at the requested address",
		usage:       "dumpaddr <host>:<port>",
		do:          dumpAddress,
	}
	commands["dumpall"] = command{
		name:         "dumpall",
		description:  "Dumps info on each node in the current ring",
		do:           dumpAll,
		joinRequired: true,
	}
}

//////////////
// Commands //
//////////////

// Lists known aliases for commands
func listCommands(_ string) error {
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

	return nil
}

// Quit gracefully and offload data to other nodes
func quit(_ string) error {
	fmt.Println("Quitting...")
	os.Exit(0)
	return nil
}

// Change port to listen on, can't be done after joining
func changePort(p string) error {
	if !joined {
		newPort, err := strconv.Atoi(p)
		if err != nil {
			return fmt.Errorf("bad port: %v", err)
		}
		fmt.Printf("Listening port changed from %d to %d\n", localPort, newPort)
		localPort = newPort
	} else {
		return errors.New("can't change port. already listening")
	}
	return nil
}

func ping(inputAddress string) error {
	address, err := validateAddress((inputAddress))
	if err != nil {
		return fmt.Errorf("bad address: %v", err)
	}
	fmt.Printf("Attempting to ping %s...\n", address)
	var success bool
	if err := call(address, "NodeActor.Ping", None{}, &success); err != nil {
		return fmt.Errorf("ping: %v", err)
	}
	fmt.Println("Success")
	return nil
}

func create(_ string) error {
	if !joined {
		var err error
		if localNode, err = createRing(); err != nil {
			return fmt.Errorf("creating ring: %v", err)
		}
		// Successful creation of new ring
		joined = true
	} else {
		return errors.New("can't create ring. already part of a ring")
	}
	return nil
}

func join(inputAddress string) error {
	if !joined {
		address, err := validateAddress(inputAddress)
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		if localNode, err = joinRing(address); err != nil {
			return fmt.Errorf("joining ring: %v", err)
		}
		// Successful join
		joined = true
	} else {
		return errors.New("can't join ring. already part of a ring")
	}
	return nil
}

// Dump info on local node
func dump(_ string) error {
	fmt.Println(localNode)
	return nil
}

// Dumps info on the node responsible for a key
func dumpKey(key string) error {
	return nil
}

// Dumps info on the node at the requested address
func dumpAddress(inputAddress string) error {
	address, err := validateAddress(inputAddress)
	if err != nil {
		return fmt.Errorf("bad address: %v", err)
	}
	var dump string
	if err := call(address, "NodeActor.Dump", None{}, &dump); err != nil {
		return fmt.Errorf("getting dump info: %v", err)
	}
	fmt.Println(dump)
	return nil
}

// Dumps info on each node in the current ring
func dumpAll(_ string) error {
	return nil
}

func put(input string) error {
	if words := strings.Fields(input); len(words) == 2 {
		key, value := Key(words[0]), words[1]
		fmt.Printf("Put: %s => %s\n", key, value)
		kv := KeyValue{key, value}
		// Find address to put at
		address, err := find(key.hashed(), localNode.Address)
		if err != nil {
			return fmt.Errorf("finding correct node to put at: %v", err)
		}
		// Now put it there
		if err := call(address, "NodeActor.Put", kv, &None{}); err != nil {
			return fmt.Errorf("putting: %v", err)
		}
		fmt.Println("successful put: ", kv)
	} else {
		return errors.New("too many values: <key> <value>")
	}
	return nil
}

func get(input string) error {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(words[0])
		fmt.Printf("Get item with key: %s\n", key)
		// Find address to get from
		address, err := find(key.hashed(), localNode.Address)
		if err != nil {
			return fmt.Errorf("finding correct node to get from: %v", err)
		}
		// Now get the value
		var value string
		if err := call(address, "NodeActor.Get", key, &value); err != nil {
			return fmt.Errorf("getting: %v", err)
		}
		fmt.Println(KeyValue{key, value})
	} else {
		fmt.Println("Too many values: <key>")
	}
	return nil
}

func deleteKey(input string) error {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(words[0])
		fmt.Printf("Delete item with key: %s", key)
		// Find address to delete from
		address, err := find(key.hashed(), localNode.Address)
		if err != nil {
			return fmt.Errorf("finding correct node to delete from: %v", err)
		}
		// Now delete the value
		var value string
		if err := call(address, "NodeActor.Delete", key, &value); err != nil {
			return fmt.Errorf("deleting: %v", err)
		}
		fmt.Printf("Successfully deleted item with key: %s, and value %s", key, value)
	} else {
		fmt.Println("Too many values: <key>")
	}
	return nil
}

func putRandom(count string) error {
	return nil
}
