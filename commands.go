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
		name        string
		description string
		usage       string
		do          func(string) error
	}
)

var (
	commands   map[string]command // Map of command aliases to commands
	ansiColors map[string]string  // ANSI colors to code map

	notJoinedMsg = "must join a ring before querying nodes"
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
		description: "Delete a key and its associated value",
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

func ping(address string) error {
	if joined {
		addr, err := validateAddress((address))
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		fmt.Printf("Attempting to ping %s...\n", address)
		var success bool
		if err := call(addr, "NodeActor.Ping", None{}, &success); err != nil {
			return fmt.Errorf("ping: %v", err)
		}
		fmt.Println("Success")
	} else {
		return errors.New(notJoinedMsg)
	}
	return nil
}

func create(_ string) error {
	if !joined {
		if err := startNode(); err != nil {
			return fmt.Errorf("create error: %v", err)
		}
		joined = true
	} else {
		return errors.New("can't create ring. already part of a ring")
	}
	return nil
}

func join(address string) error {
	if !joined {
		addr, err := validateAddress(address)
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		if err := startNode(); err != nil {
			return fmt.Errorf("create error: %v", err)
		}
		joined = true
		fmt.Println(addr, addr.hashed())
	} else {
		return errors.New("can't join ring. already part of a ring")
	}
	return nil
}

func dump(_ string) error {
	if joined {
		fmt.Println(localNode)
	} else {
		return errors.New(notJoinedMsg)
	}
	return nil
}

// Dumps info on the node responsible for a key
func dumpKey(key string) error {
	return nil
}

// Dumps info on the node at the requested address
func dumpAddress(address string) error {
	return nil
}

// Dumps info on each node in the current ring
func dumpAll(_ string) error {
	return nil
}

func put(data string) error {
	if words := strings.Fields(data); len(words) == 3 {
		key, value := Key(words[0]), words[1]
		address, err := validateAddress(words[2])
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		fmt.Printf("Put: %s => %s\n", key, value)
		kv := KeyValue{key, value}
		// var address *Address
		// err := call(localAddress, "NodeActor.Find", struct{key, localAddress}, address)
		// if err != nil {
		// 	return fmt.Errorf("put: %v", err)
		// }

		if err := call(address, "NodeActor.Put", kv, &None{}); err != nil {
			return fmt.Errorf("put: %v", err)
		}
		fmt.Println("successful put: ", kv)
	} else {
		return errors.New("too many values: <key> <value>")
	}
	return nil
}

func get(input string) error {
	if words := strings.Fields(input); len(words) == 2 {
		key := Key(words[0])
		address, err := validateAddress(words[1])
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		fmt.Printf("Get item with key: %s\n", key)
		// err := call(localAddress, "NodeActor.Find", struct{key, localAddress}, address)
		// if err != nil {
		// 	fmt.Errorf("get finding: %v", err)
		// }
		var value string
		if err := call(address, "NodeActor.Get", key, &value); err != nil {
			return fmt.Errorf("get getting: %v", err)
		}
		fmt.Println(KeyValue{key, value})
	} else {
		fmt.Println("Too many values: <key>")
	}
	return nil
}

func deleteKey(input string) error {
	if words := strings.Fields(input); len(words) == 2 {
		key := Key(words[0])
		address, err := validateAddress(words[1])
		if err != nil {
			return fmt.Errorf("bad address: %v", err)
		}
		fmt.Printf("Delete item with key: %s", key)
		// var address *Address
		// err := call(localAddress, "NodeActor.Find", struct{key, localAddress}, address)
		// if err != nil {
		// 	fmt.Errorf("delete finding: %v", err)
		// }
		var value string
		if err := call(address, "NodeActor.Delete", key, &value); err != nil {
			return fmt.Errorf("delete deleting: %v", err)
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
