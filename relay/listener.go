package relay

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
)

// Listener implements the net.Listener interface in such a way that servers can
// easily be setup to communicate with the relay.
type Listener struct {
	clients map[string]*Conn
	msgs    chan *Message
	lock    sync.Mutex
	wg      sync.WaitGroup
	conn    net.Conn
	dec     *json.Decoder
	in      chan net.Conn
	close   chan struct{}
}

// Dial connects to a tcprelay server using the given addr:port. It acts as a
// net.Listener by handling messages from a relay server. It also returns the
// relay information.
func Dial(addr string) (*Listener, string, error) {
	l := &Listener{
		clients: make(map[string]*Conn),
		msgs:    make(chan *Message),
		in:      make(chan net.Conn),
		close:   make(chan struct{}),
	}
	// Make the connection
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	l.conn = conn
	// Get the first message.
	l.dec = json.NewDecoder(conn)
	msg := &Message{}
	err = l.dec.Decode(msg)
	if err != nil {
		l.conn.Close()
		return nil, "", errors.New("failed to decode relay message")
	} else if msg.Type != MessageTypeRelay {
		return nil, "", errors.New("relay message wasn't the first message")
	}
	// Startup the goroutines that listen for messages and return.
	go l.handleMessagesToRelay()
	go l.handleMessagesFromRelay()
	return l, string(msg.Data), nil
}

func (l *Listener) handleMessagesToRelay() {
	l.wg.Add(1)
	defer l.wg.Done()
	for {
		// Wait for messages.
		msg, ok := <-l.msgs
		if !ok {
			return
		}
		// If we got a close mesage, we need to remove it from our client list.
		if msg.Type == MessageTypeClose {
			l.lock.Lock()
			c := l.clients[msg.RemoteAddr]
			if c == nil {
				log.Printf("no connection to %v, not closed.", msg.RemoteAddr)
				l.lock.Unlock()
				continue
			}
			delete(l.clients, msg.RemoteAddr)
			l.lock.Unlock()
		}
		// Encode the messsage and write it to our relay.
		b, err := json.Marshal(msg)
		if err != nil {
			log.Printf("marshalling relay %v: %v", msg, err)
			continue
		}
		if _, err := l.conn.Write(b); err != nil {
			log.Printf("writing %v: %v", msg, err)
			return
		}
	}
}

func (l *Listener) handleMessagesFromRelay() {
	l.wg.Add(1)
	defer l.wg.Done()
	msg := &Message{}
	for {
		// Get the next message.
		err := l.dec.Decode(msg)
		if err != nil {
			log.Fatalln("getting relay:", err)
			continue
		}
		switch msg.Type {
		case MessageTypeConnect:
			// Create a new client.
			c, err := NewClient(msg.LocalAddr, msg.RemoteAddr, l.msgs)
			if err != nil {
				log.Printf("making new connection %v: %v", msg, err)
				continue
			}
			// Add it to our mapping and notify the listener.
			l.lock.Lock()
			l.clients[msg.RemoteAddr] = c
			l.lock.Unlock()
			l.in <- c
		case MessageTypeData:
			// Send data to the client.
			l.lock.Lock()
			c := l.clients[msg.RemoteAddr]
			l.lock.Unlock()
			if c == nil {
				log.Printf("no connection to %v, data not sent.", msg.RemoteAddr)
				continue
			}
			c.Data(msg.Data)
		case MessageTypeClose:
			l.lock.Lock()
			c := l.clients[msg.RemoteAddr]
			l.lock.Unlock()
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

// Accept implements the net.Conn interface. New connections from the relay will
// result in this returning a new compatible net.Conn.
func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.close:
		return nil, errors.New("closed")
	case conn := <-l.in:
		return conn, nil
	}
}

// Close stops this Listener and will signal Accept() to stop if it's waiting.
func (l *Listener) Close() error {
	close(l.close)
	l.wg.Wait()
	return nil
}

// Addr implements the net.Conn interface. It currently returns the address of
// the connection to the relay.
func (l *Listener) Addr() net.Addr {
	return l.conn.RemoteAddr()
}
