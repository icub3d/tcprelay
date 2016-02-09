// Package relay provides tools for using tcprelay. It contains the messaging
// infrastructure and some othe standard librarys net interfaces for interacting
// with the relay as if it were a real TCP connection.
package relay

import "fmt"

// MessageType is the message type. It may simply signal something or help
// unmarshal the rest of the message to it's useful parts.
type MessageType int

// String returns the string representation of the given type.
func (t MessageType) String() string {
	switch t {
	case MessageTypeRelay:
		return "relay"
	case MessageTypeStop:
		return "stop"
	case MessageTypeConnect:
		return "connect"
	case MessageTypeData:
		return "data"
	case MessageTypeClose:
		return "close"
	}
	return ""
}

const (
	// MessageTypeRelay is the first message the relay sends to the server. It's
	// data is a utf8 byte-encoded string (e.g. string(m.Data)) that contains the
	// addr:port that clients can use to connect.
	MessageTypeRelay MessageType = iota

	// MessageTypeStop is a signal from the server that the relay should shutdown
	// relaying for this server.
	MessageTypeStop

	// MessageTypeConnect is a signal from the relay to the server that a new
	// connection is being made. The RemoteAddr and LocalAddr will contain the
	// client information and should be used for future communication to the
	// client.
	MessageTypeConnect

	// MessageTypeData is how the server and relay transfer data to and from
	// clients. The RemoteAddr and LocalAddr should be filled and the Data contains
	// the new data to process.
	MessageTypeData

	// MessageTypeClose is how the server and relay signal that a client is or
	// should be closed. The RemoteAddr and LocalAddr should contain the client
	// information that should be closed.
	MessageTypeClose
)

// Message is a generic message that the servers and clients use to communicate.
// If the message is from or for a client, the remote and local addresses should
// be filled.
type Message struct {
	Type       MessageType
	RemoteAddr string
	LocalAddr  string
	Data       []byte
}

// String returns a human readable version of this message.
func (m *Message) String() string {
	s := string(m.Data)
	if len(s) > 20 {
		s = s[:20]
	}
	return fmt.Sprintf("[%v %v %v %v]",
		m.Type, m.RemoteAddr, m.LocalAddr, s)
}
