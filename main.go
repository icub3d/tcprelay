// Program tcprelay relays TCP traffic between clients and servers.
package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	// These are the command-line arguments.
	addr      string
	ports     string
	saddr     string
	low, high int

	// the ports currently in use by servers.
	usedPorts = map[int]bool{}
	upLock    = sync.Mutex{}

	// ErrInvalidPortRange is returned when parsing the ports command-line
	// argument fails.
	ErrInvalidPortRange = errors.New("invalid port range")
)

func init() {
	flag.StringVar(&addr, "addr", ":8000",
		"the addr:port upon which servers communicate with this relay.")
	flag.StringVar(&ports, "ports", ":8001-9000",
		"the addr and port range (inclusive) wherein servers will be assigned relay ports.")
}

func main() {
	// Parse the args and make sure the range is valid.
	flag.Parse()
	var err error
	saddr, low, high, err = parsePorts(ports)
	if err != nil {
		log.Fatalf("invalid port range: %v", ports)
	}
	log.Printf("addr: %v, port range: %v:%v-%v", addr, saddr, low, high)

	// Start listening for new servers.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("unable to start server: %v", err)
	}
	for {
		// TODO - should probably exit cleanly.
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error during Accept(): %v", err)
		}
		go newServer(conn)
	}
}

// parsePorts splits up the given string into it's address, low port, and high
// port parts.
func parsePorts(ports string) (string, int, int, error) {
	parts := strings.SplitN(ports, ":", 2)
	if len(parts) < 2 {
		return "", 0, 0, ErrInvalidPortRange
	}
	addr := parts[0]
	parts = strings.SplitN(parts[1], "-", 2)
	if len(parts) < 2 {
		return "", 0, 0, ErrInvalidPortRange
	}
	low, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", 0, 0, ErrInvalidPortRange
	}
	high, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, ErrInvalidPortRange
	}
	return addr, low, high, nil
}

// findUnusedPort finds a port not in use in the port range given on the command
// line.
func findUnusedPort() int {
	upLock.Lock()
	defer upLock.Unlock()
	port := -1
	for x := low; x <= high; x++ {
		if !usedPorts[x] {
			port = x
			break
		}
	}
	usedPorts[port] = true
	return port
}

// releasePort removes the given port from the used ports so new server
// connections can use it.
func releasePort(port int) {
	upLock.Lock()
	defer upLock.Unlock()
	delete(usedPorts, port)
}
