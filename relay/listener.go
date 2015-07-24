package relay

import (
	"errors"
	"net"
)

// Listener implements the net.Listener interface in such a way that servers can
// easily be setup to communicate with the relay.
type Listener struct {
	in    chan net.Conn
	close chan struct{}
}

// NewListener creates a new Listener.
func NewListener() *Listener {
	l := &Listener{
		in:    make(chan net.Conn),
		close: make(chan struct{}),
	}
	return l
}

// NewConn queues of the given conn for Accept() and waits until Accept() has
// picked it up.
func (l *Listener) NewConn(conn net.Conn) {
	l.in <- conn
}

// Accept implements the net.Conn interface. New connections are received from
// NewConn().
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
	return nil
}

// Addr implements the net.Conn interface. It currently returns nil.
func (l *Listener) Addr() net.Addr {
	return nil
}
