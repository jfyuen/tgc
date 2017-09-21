package main

import (
	"log"
	"net"
	"time"
)

// Connector opens 2 connections and forwards data from one to another
type Connector struct {
	src, dst string
	interval int
}

func connect(p Pipe, interval int) {
	for {
		conn, err := net.Dial("tcp", p.addr)
		if err != nil {
			log.Printf("[error] cannot connect to %s: %v, retrying in %v seconds\n", p.addr, err, interval)
			time.Sleep(time.Duration(interval) * time.Second)
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
	go connect(InitPipe(fromCh, toCh, c.dst), c.interval)
	connect(InitPipe(toCh, fromCh, c.src), c.interval)
}
