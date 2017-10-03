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

func waitForConn(ln net.Listener, fromCh, toCh chan Message, isInNode bool) {
	addr := ln.Addr().String()
	for {
		onError := make(chan error)
		var p Pipe
		if isInNode {
			p = InNode{from: fromCh, to: toCh, err: make(chan error, 1), addr: addr}
		} else {
			p = OutNode{from: fromCh, to: toCh, err: make(chan error, 1), addr: ln.Addr().String()}
		}
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[error] accept failed: %s\n", err)
		} else {
			log.Printf("accepted new connection on %s\n", addr)
			p.Wait(conn, onError)
			<-onError
			conn.Close()
		}
	}
}

func listen(fromCh, toCh chan Message, config *tls.Config, addr string, isInNode bool) error {
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
	if isInNode {
		waitForConn(ln, fromCh, toCh, isInNode)
	} else {
		go waitForConn(ln, fromCh, toCh, isInNode)
	}
	return nil
}

// Listen on two ports and creates a custom Pipe between them
func (l Listener) Listen() error {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	if err := listen(fromCh, toCh, nil, l.from, false); err != nil {
		return err
	}
	if err := listen(toCh, fromCh, l.tlsConfig, l.to, true); err != nil {
		return err
	}
	return nil
}
