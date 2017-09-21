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

func waitForConn(ln net.Listener, p Pipe) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[error] accept failed: %s\n", err)
		} else {
			log.Printf("accepted new connection on %s\n", p.addr)
			p.Wait(conn)
			if err := <-p.receiveError; err != nil {
				conn.Close()
			}
		}
	}
}

func listen(p Pipe, block bool) error {
	addr := p.addr
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s\n", addr)
	if block {
		waitForConn(ln, p)
	} else {
		go waitForConn(ln, p)
	}
	return nil
}

// Listen on two ports and creates a custom Pipe between them
func (l Listener) Listen() error {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	if err := listen(InitPipe(fromCh, toCh, l.from), false); err != nil {
		return err
	}
	if err := listen(InitPipe(toCh, fromCh, l.to), true); err != nil {
		return err
	}
	return nil
}
