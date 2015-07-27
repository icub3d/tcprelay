package main

import (
	"flag"
	"log"
	"net/http"

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

	l, client, err := relay.Dial(relayAddr)
	if err != nil {
		log.Fatalln("connecting to relay relay:", err)
	}
	log.Println("client connection:", client)

	s := &http.Server{
		Handler: http.FileServer(http.Dir(dir)),
	}
	log.Println(s.Serve(l))
}
