package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/icub3d/tcprelay/relay"
)

// Server contains the information about a connecting server. It should be
// created with the newServer function.
type server struct {
	port     int
	conn     net.Conn
	listener net.Listener
	clients  map[string]*client
	lock     sync.Mutex
	toServer chan *relay.Message
	close    chan struct{}
	wg       sync.WaitGroup
}

// newServer sets up a new server connection. It will communicate with the
// server what port it's publishing and then forward any tcp traffic.
func newServer(conn net.Conn) {
	var err error
	s := &server{
		conn:     conn,
		clients:  make(map[string]*client),
		toServer: make(chan *relay.Message),
		close:    make(chan struct{}),
	}
	// Find an unused port.
	s.port = findUnusedPort()
	if s.port == -1 {
		// We didn't find one, notify the server and exit!
		log.Println("didn't find an open port for server:", conn.RemoteAddr())
		return
	}
	addr := fmt.Sprintf("%v:%v", saddr, s.port)
	// create the listener for clients for this server.
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Printf("unable to listen for %v: %v", conn.RemoteAddr(), err)
		return
	}
	// Start up the server goroutines.
	go s.handleMessagesFromServer()
	go s.handleMessagesToServer()
	// Send the relay relay.
	msg := &relay.Message{
		Type: relay.MessageTypeRelay,
		Data: []byte(addr),
	}
	if !s.Send(msg) {
		return
	}
	// Start listeneing for clients.
	go s.listen()
}

// String returns the RemoteAddr for this server.
func (s *server) String() string {
	return s.conn.RemoteAddr().String()
}

// Close closes all the open connections and waits for all the goroutines to
// finish. It then release the port being used by this server.
func (s *server) Close() {
	// Close the server connection and client listener.
	close(s.close)
	s.listener.Close()
	// Close all of the clients.
	s.lock.Lock()
	for _, c := range s.clients {
		if err := c.Close(); err != nil {
			log.Printf("[%v] closing %v: %v", s, c, err)
		}
	}
	s.lock.Unlock()
	// Wait for our goroutines to finish and then release the port.
	s.wg.Wait()
	releasePort(s.port)
}

// Send sends the given relay to the server. It returns true if successful. If
// the server was closed while trying to send, false is returned.
func (s *server) Send(msg *relay.Message) bool {
	select {
	case s.toServer <- msg:
		return true
	case <-s.close:
		return false
	}
}

// handleMessagesFromServer reads messages from the server and handles them
// appropriately.
func (s *server) handleMessagesFromServer() {
	// TODO the WaitGroup is sort of funky here with the Close() we can probably
	// tighten this up a bit.
	s.wg.Add(1)
	dec := json.NewDecoder(s.conn)
	for {
		// Get a relay.
		msg := &relay.Message{}
		err := dec.Decode(msg)
		if err != nil {
			log.Printf("[%v] decoding relay : %v", s, err)
			s.wg.Done()
			s.Close()
			return
		}
		// Do something based on the relay.
		switch msg.Type {
		case relay.MessageTypeStop:
			s.wg.Done()
			s.Close()
			return
		case relay.MessageTypeData:
			c := s.getClient(msg.RemoteAddr)
			if c == nil {
				log.Printf("[%v] data not sent - no client: %v", s, msg.RemoteAddr)
				continue
			}
			if err := c.Send(msg.Data); err != nil {
				log.Printf("[%v] sending to %v: %v", s, c, err)
			}
		case relay.MessageTypeClose:
			c := s.getClient(msg.RemoteAddr)
			if c == nil {
				log.Printf("[%v] unabel to close - no client: %v", s, msg.RemoteAddr)
				continue
			}
			if err := c.Close(); err != nil {
				log.Printf("[%v] closing %v: %v", s, c, err)
			}
		default:
			log.Printf("[%v] unexpected relay: %v", s, msg)
		}
	}
}

// handleMessagesToServer loops reading from the channel for messages that
// should be sent to the server. It sends those messages over the connection.
func (s *server) handleMessagesToServer() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		// Get the next relay or exit.
		var msg *relay.Message
		var ok bool
		select {
		case msg, ok = <-s.toServer:
			if !ok {
				return
			}
		case <-s.close:
			return
		}
		// Send the relay.
		d, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[%v] marshalling relay: %v", s, err)
			continue
		}
		_, err = s.conn.Write(d)
		if err != nil {
			log.Printf("[%v] sending relay : %v", s, err)
			break
		}
	}
}

// getClient returns the client mapped to the given remote address.
func (s *server) getClient(addr string) *client {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.clients[addr]
}

// listen loops Accept()ing for client connections. When it gets
// one, it creates a new client struct and adds it to our client table.
func (s *server) listen() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("[%v] accepting: %v", s, err)
			break
		}
		// Send the new connection relay.
		msg := &relay.Message{
			Type:       relay.MessageTypeConnect,
			RemoteAddr: conn.RemoteAddr().String(),
			LocalAddr:  conn.LocalAddr().String(),
		}
		if !s.Send(msg) {
			break
		}
		// Setup the new client and add it to our table.
		c := newClient(conn, s)
		s.lock.Lock()
		s.clients[conn.RemoteAddr().String()] = c
		s.lock.Unlock()
	}
}
