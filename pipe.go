package main

import (
	"bytes"
	"io"
	"log"
	"net"
)

// Pipe reads and writes on a connection, results are sent over channels
type Pipe struct {
	from, to     chan []byte
	err          chan error
	receiveError chan error
	sendError    chan error
}

func (p Pipe) receive(conn net.Conn, port string) {
	b := make([]byte, 4096)
	for {
		n, err := conn.Read(b)
		log.Printf("read %d on %s with err: %v\n", n, port, err)
		if err != nil {
			p.err <- err
			if p.receiveError != nil {
				p.receiveError <- err
			}
			return
		}
		if n > 0 {
			data := make([]byte, n)
			copy(data, b[:n])
			p.to <- data
		}
	}
}

func (p Pipe) send(conn net.Conn, port string) {

	for {
		select {
		case data := <-p.from:
			log.Printf("send %v bytes on %s\n", len(data), port)
			buf := bytes.NewBuffer(data)
			if _, err := io.Copy(conn, buf); err != nil {
				log.Printf("[error] could not write %v bytes on %s", len(data), port)
				p.from <- data
				if p.sendError != nil {
					p.sendError <- err
				}
				return
			}
		case err := <-p.err:
			log.Printf("[error] received %v on %s", err, port)
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p Pipe) Wait(conn net.Conn, port string) {
	go p.receive(conn, port)
	go p.send(conn, port)
}

// InitPipe creates a Pipe with some preinitialized channels
func InitPipe(from, to chan []byte) Pipe {
	return Pipe{from: from, to: to, err: make(chan error), receiveError: make(chan error)}
}
