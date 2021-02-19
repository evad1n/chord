package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
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
func validateAddress(address string) bool {
	// Regex for <IPv4>:<PORT>
	matched, _ := regexp.Match(`^(?:\d+\.){3}\d+:(?:\d?){4}\d$`, []byte(address))
	return matched
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func (kv KeyValue) String() string {
	hex := fmt.Sprintf("%040x", kv.Key)
	return fmt.Sprintf("%s.. (%s)", hex[:8], string(kv.Value))
}
