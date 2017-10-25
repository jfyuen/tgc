package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// Connector opens 2 connections and forwards data from one to another
type Connector struct {
	src, dst  string
	interval  int
	tlsConfig *tls.Config
}

func connect(fromCh, toCh chan Message, interval int, config *tls.Config, addr string, isInNode bool) {
	for {
		var conn net.Conn
		var err error
		if config != nil {
			conn, err = tls.Dial("tcp", addr, config)
		} else {
			conn, err = net.Dial("tcp", addr)
		}
		if err != nil {
			errorLog.Printf("cannot connect to %s: %v, retrying in %v seconds\n", addr, err, interval)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		msg := fmt.Sprintf("connected to %s\n", addr)
		if config != nil {
			msg = "securely " + msg
		}
		debugLog.Print(msg)
		ctx, cancel := context.WithCancel(context.Background())
		var p Pipe
		if isInNode {
			p = InNode{from: fromCh, to: toCh, cancel: cancel, addr: addr}
		} else {
			p = OutNode{from: fromCh, to: toCh, cancel: cancel, addr: addr}
		}

		p.Wait(ctx, conn)
		<-ctx.Done()
		conn.Close()
	}
}

// Connect the two destinations using a custom Pipe
func (c Connector) Connect() {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	go connect(fromCh, toCh, c.interval, c.tlsConfig, c.dst, true)
	connect(toCh, fromCh, c.interval, nil, c.src, false)
}
