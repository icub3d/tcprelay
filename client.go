package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/icub3d/tcprelay/relay"
)

// client is a connection from the other side of the relay. Clients connections
// are made through the server struct.
type client struct {
	conn   net.Conn
	server *server
	closed bool
	lock   sync.Mutex
	wg     sync.WaitGroup
}

// newClient creates a new client for the given net.Conn. It sends new data from
// the client to the given channel. When the client should be closed from the
// server side, Close() should be called.
func newClient(conn net.Conn, server *server) *client {
	c := &client{conn: conn, server: server}
	go c.run()
	return c
}

// String returns the LocalAddr/RemoteAddr for this server.
func (c *client) String() string {
	return fmt.Sprintf("[%v %v]", c.conn.LocalAddr().String(),
		c.conn.RemoteAddr().String())
}

// Close disconnects the client connection.
func (c *client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.closed = true
	err := c.conn.Close()
	c.wg.Wait()
	return err
}

// Send sends the given message to this client.
func (c *client) Send(p []byte) error {
	_, err := c.conn.Write(p)
	return err
}

func (c *client) run() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		// Read a message.
		buf := make([]byte, 4096)
		n, err := c.conn.Read(buf)
		msg := &relay.Message{
			RemoteAddr: c.conn.RemoteAddr().String(),
			LocalAddr:  c.conn.LocalAddr().String(),
		}
		if err != nil {
			// If we didn't close via Close(), we should signal the server and then
			// Close() ourselves.
			closed := false
			c.lock.Lock()
			closed = c.closed
			c.lock.Unlock()
			if !closed {
				msg.Type = relay.MessageTypeClose
				c.server.Send(msg)
			}
			return
		}
		// Send the data to the server.
		msg.Type = relay.MessageTypeData
		msg.Data = buf[:n]
		c.server.Send(msg)
	}
}
