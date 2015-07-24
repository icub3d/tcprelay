package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/icub3d/tcprelay/relay"
)

// TODO we should probably handle shutdowns properly, right now the world just
// stops abruptly. Use a signalhandler to catch the signal, close the listener
// so we don't accept anymore, close all the clients, and send the stop to the
// relay.

var (
	relayAddr string
	dir       string
)

func init() {
	flag.StringVar(&relayAddr, "relay", "localhost:8000",
		"the addr:port of the relay server.")
	flag.StringVar(&dir, "dir", ".",
		"the directory to serve.")
}

func main() {
	flag.Parse()

	// Make the connection.
	conn, err := net.Dial("tcp", relayAddr)
	if err != nil {
		log.Fatalln("connecting to relay:", err)
	}

	// The first relay should be our relay relay.
	dec := json.NewDecoder(conn)
	msg := &relay.Message{}
	err = dec.Decode(msg)
	if msg.Type != relay.MessageTypeRelay {
		log.Fatalln("unexpected relay:", msg)
	}
	log.Println("client connection:", string(msg.Data))

	// setup our fake listener and the http server.
	msgs := make(chan *relay.Message)
	listener := relay.NewListener()
	s := &http.Server{
		Handler: http.FileServer(http.Dir(dir)),
	}
	go s.Serve(listener)

	// Setup a gorouting that gets messages and writes them to the relay.
	clients := make(map[string]*relay.Conn)
	lock := new(sync.Mutex)
	go func() {
		for {
			// Wait for messages.
			msg, ok := <-msgs
			if !ok {
				return
			}
			// If we got a close relay, we need to remove it from our client list.
			if msg.Type == relay.MessageTypeClose {
				lock.Lock()
				c := clients[msg.RemoteAddr]
				if c == nil {
					log.Printf("not connection to %v, not closed.", msg.RemoteAddr)
					lock.Unlock()
					continue
				}
				delete(clients, msg.RemoteAddr)
				lock.Unlock()
			}
			// Encode the relay and write it to our relay.
			b, err := json.Marshal(msg)
			if err != nil {
				log.Printf("marshalling relay %v: %v", msg, err)
				continue
			}
			if _, err := conn.Write(b); err != nil {
				log.Printf("writing %v: %v", msg, err)
				return
			}
		}
	}()

	// Now the main loop simply reads from the relay and handles the messages
	// appropriately.
	for {
		err = dec.Decode(msg)
		if err != nil {
			log.Fatalln("getting relay:", err)
			continue
		}
		switch msg.Type {
		case relay.MessageTypeConnect:
			// Create a new client.
			c, err := relay.NewConn(msg.LocalAddr, msg.RemoteAddr, msgs)
			if err != nil {
				log.Printf("making new connection %v: %v", msg, err)
				continue
			}
			// Add it to our mapping and notify the listener.
			lock.Lock()
			clients[msg.RemoteAddr] = c
			lock.Unlock()
			listener.NewConn(c)
		case relay.MessageTypeData:
			// Send data to the client.
			lock.Lock()
			c := clients[msg.RemoteAddr]
			lock.Unlock()
			if c == nil {
				log.Printf("no connection to %v, data not sent.", msg.RemoteAddr)
				continue
			}
			c.Data(msg.Data)
		case relay.MessageTypeClose:
			lock.Lock()
			c := clients[msg.RemoteAddr]
			lock.Unlock()
			if c == nil {
				log.Printf("not connection to %v, not closed.", msg.RemoteAddr)
				continue
			}
			c.Close() // This will notify the relay and remove it from the mapping.
		default:
			log.Printf("unrecognized relay: %v", msg)
		}
	}
}
