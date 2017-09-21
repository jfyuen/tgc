package main

import (
	"log"
	"net"
	"time"
)

// Connector opens 2 connections and forwards data from one to another
type Connector struct {
	src, dst string
}

func connect(addr string, p Pipe) {
	p.addr = addr
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			wait := 5
			log.Printf("[error] cannot connect to %s: %v, retrying in %v seconds\n", addr, err, wait)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("connected for %s", addr)
		p.Wait(conn)
		err = <-p.receiveError
		conn.Close()
	}
}

// Connect the two destinations using a custom Pipe
func (c Connector) Connect() {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	go connect(c.dst, InitPipe(fromCh, toCh))
	connect(c.src, InitPipe(toCh, fromCh))
}
