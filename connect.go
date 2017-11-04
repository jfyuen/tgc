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
	reconnect bool // Automatically reconnect to the local server
	tlsConfig *tls.Config
}

func (c Connector) connectOutNode(fromCh, toCh chan Message) {
	addr := c.src
	for {
		bufCh := make(chan Message)
		if !c.reconnect {
			msg := <-fromCh // block until data are received from the remote node
			go func(m Message) {
				bufCh <- msg
			}(msg)
		}
		conn, err := net.Dial("tcp", addr)

		if err != nil {
			errorLog.Printf("cannot connect to %s: %v, retrying in %v seconds\n", addr, err, c.interval)
			time.Sleep(time.Duration(c.interval) * time.Second)
			continue
		}

		msg := fmt.Sprintf("connected to %s\n", addr)
		debugLog.Print(msg)
		ctx, cancel := context.WithCancel(context.Background())

		// Use bufCh as a buffer to block automatically reconnecting
		go func(ctx context.Context) {
			for {
				select {
				case m := <-fromCh:
					bufCh <- m
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
		p := OutNode{from: bufCh, to: toCh, cancel: cancel, addr: addr}

		p.Wait(ctx, conn)
		<-ctx.Done()
		conn.Close()
	}
}

func (c Connector) connectInNode(fromCh, toCh chan Message) {
	addr := c.dst
	for {
		var conn net.Conn
		var err error
		if c.tlsConfig != nil {
			conn, err = tls.Dial("tcp", addr, c.tlsConfig)
		} else {
			conn, err = net.Dial("tcp", addr)
		}

		if err != nil {
			errorLog.Printf("cannot connect to %s: %v, retrying in %v seconds\n", addr, err, c.interval)
			time.Sleep(time.Duration(c.interval) * time.Second)
			continue
		}
		msg := fmt.Sprintf("connected to %s\n", addr)
		if c.tlsConfig != nil {
			msg = "securely " + msg
		}
		debugLog.Print(msg)

		ctx, cancel := context.WithCancel(context.Background())
		p := InNode{from: fromCh, to: toCh, cancel: cancel, addr: addr}
		p.Wait(ctx, conn)
		<-ctx.Done()
		conn.Close()
	}
}

// Connect the two destinations using a custom Pipe over channel
func (c Connector) Connect() {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	go c.connectInNode(fromCh, toCh)
	c.connectOutNode(toCh, fromCh)
}
