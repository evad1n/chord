package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

// Get local IP address
func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// Wrap some text in an ansi code
func ansiWrap(text string, code string) string {
	return fmt.Sprintf("%s%s\x1b[0m", code, text)
}

// Centers text in the middle of a column of size {size}
func centerText(text string, size int, fill rune) string {
	if &fill == nil {
		fill = ' '
	}
	size -= len(text)
	front := size / 2
	return strings.Repeat(string(fill), front) + text + strings.Repeat(string(fill), size-front)
}

// Validate an address (host IP + port)
func validateAddress(address string) (Address, error) {
	// Regex for <IPv4>:<PORT>
	matched, _ := regexp.Match(`^(?:\d+\.){3}\d+:(?:\d?){4}\d$`, []byte(address))
	if matched {
		return Address(address), nil
	}
	return Address(address), errors.New("invalid address format: <host>:<port>")
}
