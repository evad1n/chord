package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type (
	command struct {
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

// Displaying command widths
var (
	nameWidth        int
	descriptionWidth int
	usageWidth       int
	totalWidth       int
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
		description: "List all commands and descriptions",
		do:          listCommands,
	}
	commands["quit"] = command{
		description: "Quit and offload node data gracefully",
		do:          quit,
	}
	commands["port"] = command{
		description: "Change the listening port",
		usage:       "port <number>",
		do:          changePort,
	}
	commands["getaddr"] = command{
		description: "Get the current node address",
		do: func(_ string) error {
			fmt.Println(localNode.Address)
			return nil
		},
		joinRequired: true,
	}
	commands["ping"] = command{
		description: "Ping another node",
		usage:       "ping <host>:<port>",
		do:          ping,
	}
	commands["create"] = command{
		description: "Create and join a new chord ring",
		do:          create,
	}
	commands["join"] = command{
		description: "Join a chord ring from a known node address",
		usage:       "join <host>:<port>",
		do:          join,
	}
	commands["put"] = command{
		description:  "Add a key/value pair to the database",
		usage:        "put <key> <value>",
		do:           put,
		joinRequired: true,
	}
	commands["get"] = command{
		description:  "Get the value of a key",
		usage:        "get <key>",
		do:           get,
		joinRequired: true,
	}
	commands["delete"] = command{
		description:  "Delete a key and its associated value",
		usage:        "delete <key>",
		do:           deleteKey,
		joinRequired: true,
	}
	commands["putrandom"] = command{
		description:  "Add random data items to the database",
		usage:        "putrandom <num_items>",
		do:           putRandom,
		joinRequired: true,
	}
	// Information/debugging
	commands["dump"] = command{
		description:  "Dumps current node information",
		do:           dumpCurrent,
		joinRequired: true,
	}
	commands["dumpkey"] = command{
		description:  "Dumps info on the node responsible for a key",
		usage:        "dumpkey <key>",
		do:           dumpKey,
		joinRequired: true,
	}
	commands["dumpaddr"] = command{
		description: "Dumps info on the node at the requested address",
		usage:       "dumpaddr <host>:<port>",
		do:          dumpAddress,
	}
	commands["dumpall"] = command{
		description:  "Dumps info on each node in the current ring",
		do:           dumpAll,
		joinRequired: true,
	}

	// Set display variables
	for name, cmd := range commands {
		if len(cmd.description) > descriptionWidth {
			descriptionWidth = len(cmd.description)
		}
		if len(name) > nameWidth {
			nameWidth = len(name)
		}
		if len(cmd.usage) > usageWidth {
			usageWidth = len(cmd.usage)
		}
	}
	totalWidth = nameWidth + descriptionWidth + usageWidth + 8
}

//////////////
// Commands //
//////////////

// Lists known aliases for commands
func listCommands(_ string) error {
	var w strings.Builder

	w.WriteString(fmt.Sprintf("+%s+\n", strings.Repeat("-", totalWidth)))
	w.WriteString(fmt.Sprintf("|%s|\n", centerText("COMMANDS LIST", totalWidth, ' ')))
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth+2),
		strings.Repeat("-", descriptionWidth+2),
		strings.Repeat("-", usageWidth+2),
	))
	w.WriteString(fmt.Sprintf(
		"| %s | %s | %s |\n",
		centerText("NAME", nameWidth, ' '),
		centerText("DESCRIPTION", descriptionWidth, ' '),
		centerText("USAGE", usageWidth, ' '),
	))
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth+2),
		strings.Repeat("-", descriptionWidth+2),
		strings.Repeat("-", usageWidth+2),
	))

	type namedCmd struct {
		name string
		command
	}
	cmds := []namedCmd{}
	for name, cmd := range commands {
		cmds = append(cmds, namedCmd{
			name,
			cmd,
		})
	}

	// Sort by name
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].name < cmds[j].name
	})

	for _, c := range cmds {
		w.WriteString(fmt.Sprintf("| %s | %s |", padText(c.name, nameWidth), padText(c.description, descriptionWidth)))
		if c.usage != "" {
			w.WriteString(fmt.Sprintf(" %-s |\n", padText(c.usage, usageWidth)))
		} else {
			w.WriteString(fmt.Sprintf(" %-s |\n", padText(c.name, usageWidth)))
		}
	}
	w.WriteString(fmt.Sprintf(
		"+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth+2),
		strings.Repeat("-", descriptionWidth+2),
		strings.Repeat("-", usageWidth+2),
	))

	fmt.Println(w.String())

	return nil
}

// Quit gracefully and offload data to other nodes
func quit(_ string) error {
	fmt.Println("Quitting...")
	if joined {
		// Offload all keys
		if localNode.Successors[0] != localNode.Address {
			if err := call(localNode.Successors[0], "NodeActor.PutAll", localNode.Data, &None{}); err != nil {
				// Will not actually quit; let user handle
				return fmt.Errorf("offloading data to successor: %v", err)
			}
			log.Println("Successfully offloaded data to successor")
		} else {
			fmt.Print(ansiWrap(`
Last node in ring
Data will be lost on quit
Are you sure? (y/n) `,
				ansiColors["yellow"],
			))
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			switch {
			case scanner.Text() == "y":
				fmt.Println("Ring terminated")
			default:
				return errors.New("quit aborted")
			}
		}
	}
	fmt.Println(ansiWrap("Goodbye!", ansiColors["cyan"]))
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
		fmt.Printf("Local Address: %s\n", localNode.Address)
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
		fmt.Printf("Local Address: %s\n", localNode.Address)
	} else {
		return errors.New("can't join ring. already part of a ring")
	}
	return nil
}

// Dump info on local node
func dumpCurrent(_ string) error {
	fmt.Println(localNode)
	return nil
}

// Dumps info on the node responsible for a key
func dumpKey(input string) error {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(words[0])
		fmt.Printf("Get item with key: %s\n", key)
		// Find address to get from
		address, err := find(key.hashed(), localNode.Address)
		if err != nil {
			return fmt.Errorf("finding node with key: %v", err)
		}
		// Get dump info
		var dump DumpReturn
		if err := call(address, "NodeActor.Dump", None{}, &dump); err != nil {
			return fmt.Errorf("getting dump info: %v", err)
		}
		fmt.Println(dump.Dump)
	} else {
		fmt.Println("Wrong number of arguments: <key>")
	}
	return nil
}

// Dumps info on the node at the requested address
func dumpAddress(inputAddress string) error {
	address, err := validateAddress(inputAddress)
	if err != nil {
		return fmt.Errorf("bad address: %v", err)
	}
	// Get dump info
	var dump DumpReturn
	if err := call(address, "NodeActor.Dump", None{}, &dump); err != nil {
		return fmt.Errorf("getting dump info: %v", err)
	}
	fmt.Println(dump.Dump)
	return nil
}

// Dumps info on each node in the current ring
func dumpAll(_ string) error {
	// First print current node
	fmt.Println("Current Node:")
	fmt.Println(localNode)

	dump := DumpReturn{
		Dump:      "",
		Successor: localNode.Successors[0],
	}
	for dump.Successor != localNode.Address {
		// Now get the value
		if err := call(dump.Successor, "NodeActor.Dump", None{}, &dump); err != nil {
			return fmt.Errorf("getting dump info: %v", err)
		}
		// Separator
		fmt.Println(strings.Repeat("=", 50) + "\n")
		fmt.Println(dump.Dump)
	}
	return nil
}

func put(input string) error {
	if words := strings.Fields(input); len(words) == 2 {
		key, value := Key(words[0]), words[1]
		fmt.Printf("Put: %s => %s\n", key, value)
		kv := KeyValue{key, value}
		if err := putOne(kv); err != nil {
			return fmt.Errorf("put error: %v", err)
		}
	} else {
		return errors.New("Wrong number of arguments: <key> <value>")
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
		fmt.Println("Wrong number of arguments: <key>")
	}
	return nil
}

func deleteKey(input string) error {
	if words := strings.Fields(input); len(words) == 1 {
		key := Key(words[0])
		fmt.Printf("Delete item with key: %s\n", key)
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
		fmt.Printf("Successfully deleted item with key: %s, and value %s\n", key, value)
	} else {
		fmt.Println("Wrong number of arguments: <key>")
	}
	return nil
}

func putRandom(input string) error {
	if count, err := strconv.Atoi(input); err != nil {
		return fmt.Errorf("bad number: %v", err)
	} else {
		for i := 0; i < count; i++ {
			kv := KeyValue{
				Key:   Key(randomString(5)),
				Value: randomString(5),
			}
			if err := putOne(kv); err != nil {
				return fmt.Errorf("put error: %v", err)
			}
		}
	}
	return nil
}

func putOne(kv KeyValue) error {
	// Find address to put at
	address, err := find(kv.Key.hashed(), localNode.Address)
	if err != nil {
		return fmt.Errorf("finding correct node to put at: %v", err)
	}
	// Now put it there
	if err := call(address, "NodeActor.Put", kv, &None{}); err != nil {
		return fmt.Errorf("putting: %v", err)
	}
	fmt.Println("successful put: ", kv)
	return nil
}
