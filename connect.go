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

func connect(p Pipe) {
	for {
		conn, err := net.Dial("tcp", p.addr)
		if err != nil {
			wait := 5
			log.Printf("[error] cannot connect to %s: %v, retrying in %v seconds\n", p.addr, err, wait)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("connected for %s", p.addr)
		p.Wait(conn)
		err = <-p.receiveError
		conn.Close()
	}
}

// Connect the two destinations using a custom Pipe
func (c Connector) Connect() {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	go connect(InitPipe(fromCh, toCh, c.dst))
	connect(InitPipe(toCh, fromCh, c.src))
}
