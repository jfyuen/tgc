package main

import (
	"log"
	"net"
	"strings"
)

// Listener will listen on 2 ports and forward data from one to another
type Listener struct {
	from, to string
}

func waitForConn(ln net.Listener, port string, p Pipe) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[error] accept failed: %s\n", err)
		} else {
			log.Printf("accepted new connection on %s\n", port)
			p.Wait(conn, port)
			if err := <-p.receiveError; err != nil {
				conn.Close()
			}
		}
	}
}

func listen(port string, pipe Pipe, block bool) error {
	if !strings.Contains(port, ":") {
		port = ":" + port
	}
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s\n", port)
	if block {
		waitForConn(ln, port, pipe)
	} else {
		go waitForConn(ln, port, pipe)
	}
	return nil
}

// Listen on two ports and creates a custom Pipe between them
func (l Listener) Listen() error {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	if err := listen(l.from, InitPipe(fromCh, toCh), false); err != nil {
		return err
	}
	if err := listen(l.to, InitPipe(toCh, fromCh), true); err != nil {
		return err
	}
	return nil
}
