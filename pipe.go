package main

import (
	"bytes"
	"io"
	"log"
)

// Pipe reads and writes on a connection, results are sent over channels
type Pipe struct {
	from, to     chan []byte
	err          chan error
	receiveError chan error
	sendError    chan error
	addr         string
}

func (p Pipe) receive(r io.Reader) {
	b := make([]byte, 4096)
	for {
		n, err := r.Read(b)
		log.Printf("read %d on %s with err: %v\n", n, p.addr, err)
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

func (p Pipe) send(w io.Writer) {

	for {
		select {
		case data := <-p.from:
			log.Printf("send %v bytes on %s\n", len(data), p.addr)
			buf := bytes.NewBuffer(data)
			if _, err := io.Copy(w, buf); err != nil {
				log.Printf("[error] could not write %v bytes on %s", len(data), p.addr)
				p.from <- data
				if p.sendError != nil {
					p.sendError <- err
				}
				return
			}
		case err := <-p.err:
			log.Printf("[error] received %v on %s", err, p.addr)
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p Pipe) Wait(rw io.ReadWriter) {
	go p.receive(rw)
	go p.send(rw)
}

// InitPipe creates a Pipe with some preinitialized channels
func InitPipe(from, to chan []byte) Pipe {
	return Pipe{from: from, to: to, err: make(chan error), receiveError: make(chan error)}
}
