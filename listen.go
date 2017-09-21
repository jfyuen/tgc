package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
)

// Listener will listen on 2 ports and forward data from one to another
type Listener struct {
	from, to  string
	tlsConfig *tls.Config
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

func listen(p Pipe, config *tls.Config, block bool) error {
	addr := p.addr
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	var ln net.Listener
	var err error
	if config != nil {
		ln, err = tls.Listen("tcp", addr, config)
	} else {
		ln, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("listening on %s\n", addr)
	if config != nil {
		msg = "securely " + msg
	}
	log.Printf(msg)
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
	if err := listen(InitPipe(fromCh, toCh, l.from), nil, false); err != nil {
		return err
	}
	if err := listen(InitPipe(toCh, fromCh, l.to), l.tlsConfig, true); err != nil {
		return err
	}
	return nil
}
