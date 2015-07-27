// Program echoserver is an example server connecting to tcprelay.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"

	"github.com/icub3d/tcprelay/relay"
)

var (
	relayAddr string
	reverse   bool
)

func init() {
	flag.StringVar(&relayAddr, "relay", "localhost:8000",
		"the addr:port of the relay server.")
	flag.BoolVar(&reverse, "reverse", false,
		"send the string back in reverse.")
}

func main() {
	flag.Parse()

	// Make the connection.
	conn, err := net.Dial("tcp", relayAddr)
	if err != nil {
		log.Fatalln("connecting to relay:", err)
	}

	// Setup a json encoder/decoder.
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	// The first message should be our relay message.
	msg := &relay.Message{}
	err = dec.Decode(msg)
	if msg.Type != relay.MessageTypeRelay {
		log.Fatalln("unexpected message:", msg)
	}
	log.Println("client connection:", string(msg.Data))

	// For the rest of the time, we simply read a message and write it back if it's
	// a Data message.
	for {
		err = dec.Decode(msg)
		if err != nil {
			log.Fatalln("getting message:", err)
		}
		if msg.Type != relay.MessageTypeData {
			// Ignore everything else. If we were interested in maintaining state, we'd
			// want to use a map of some sort to track new connections and close open
			// ones.
			continue
		}
		if reverse {
			msg.Data = Reverse(msg.Data)
		}
		err = enc.Encode(msg)
		if err != nil {
			log.Fatalln("sending message:", err)
		}
	}
}

// Reverse returns the given data reversed rune-wise.
func Reverse(d []byte) []byte {
	r := []rune(string(d))
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return []byte(string(r))
}
