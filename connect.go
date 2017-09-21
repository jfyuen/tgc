package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"
)

// Connector opens 2 connections and forwards data from one to another
type Connector struct {
	src, dst  string
	interval  int
	tlsConfig *tls.Config
}

func connect(p Pipe, interval int, config *tls.Config) {
	for {
		var conn net.Conn
		var err error
		if config != nil {
			conn, err = tls.Dial("tcp", p.addr, config)
		} else {
			conn, err = net.Dial("tcp", p.addr)
		}
		if err != nil {
			log.Printf("[error] cannot connect to %s: %v, retrying in %v seconds\n", p.addr, err, interval)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		msg := fmt.Sprintf("connected to %s\n", p.addr)
		if config != nil {
			msg = "securely " + msg
		}
		log.Print(msg)
		p.Wait(conn)
		err = <-p.receiveError
		conn.Close()
	}
}

// Connect the two destinations using a custom Pipe
func (c Connector) Connect() {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	go connect(InitPipe(fromCh, toCh, c.dst), c.interval, c.tlsConfig)
	connect(InitPipe(toCh, fromCh, c.src), c.interval, nil)
}
