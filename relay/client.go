package relay

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

// ErrNotImplemented is returned by the Conn deadline functions as they
// currently aren't implemented.
var ErrNotImplemented = errors.New("not implemented")

// Conn implements the net.Conn interface and interacts with relay servers.
type Conn struct {
	buf    bytes.Buffer
	closed bool
	cond   *sync.Cond
	msgs   chan *Message
	laddr  *net.TCPAddr
	raddr  *net.TCPAddr
}

// NewConn creates a new connection based on the given local and remote
// addresses. And data that should be sent to the relay will be sent via the
// given channel.
func NewConn(localAddr string, remoteAddr string, msgs chan *Message) (*Conn, error) {
	c := &Conn{
		cond: sync.NewCond(&sync.Mutex{}),
		msgs: msgs,
	}
	var err error
	c.laddr, err = net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return nil, err
	}
	c.raddr, err = net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Data queues up the given data for reading. Subsequent Read() commands will
// use the data.
func (c *Conn) Data(b []byte) {
	c.cond.L.Lock()
	c.buf.Write(b)
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

// Read attempts to fill b with any data in the buffer. If the buffer is empty,
// it will wait for data to be put using Data().
func (c *Conn) Read(b []byte) (int, error) {
	c.cond.L.Lock()
	for c.buf.Len() < 1 {
		c.cond.Wait()
		if c.closed {
			return 0, io.EOF
		}
	}
	defer c.cond.L.Unlock()
	return c.buf.Read(b)
}

// Write writes data to the connection.
func (c *Conn) Write(b []byte) (int, error) {
	// TODO -- error out if we've closed.
	cp := make([]byte, len(b))
	copy(cp, b)
	c.msgs <- &Message{
		Type:       MessageTypeData,
		RemoteAddr: c.raddr.String(),
		LocalAddr:  c.laddr.String(),
		Data:       cp,
	}
	return len(b), nil
}

// Close closes the connection and signals the relay that it's being closed.
func (c *Conn) Close() error {
	c.cond.L.Lock()
	c.closed = true
	c.cond.L.Unlock()
	c.cond.Broadcast()
	c.msgs <- &Message{
		Type:       MessageTypeClose,
		RemoteAddr: c.raddr.String(),
		LocalAddr:  c.laddr.String(),
	}
	return nil
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.laddr
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.raddr
}

// SetDeadline is not implemented and returns that error.
func (c *Conn) SetDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetReadDeadline is not implemented and returns that error.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetWriteDeadline is not implemented and returns that error.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return ErrNotImplemented
}
