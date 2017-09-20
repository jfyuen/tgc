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

func waitForConn(ln net.Listener, addr string, p Pipe) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[error] accept failed: %s\n", err)
		} else {
			log.Printf("accepted new connection on %s\n", addr)
			p.Wait(conn, addr)
			if err := <-p.receiveError; err != nil {
				conn.Close()
			}
		}
	}
}

func listen(addr string, pipe Pipe, block bool) error {
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s\n", addr)
	if block {
		waitForConn(ln, addr, pipe)
	} else {
		go waitForConn(ln, addr, pipe)
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
