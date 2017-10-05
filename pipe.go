package main

import (
	"bytes"
	"io"
)

// Pipe reads and writes on a connection, results are sent over channels
type Pipe interface {
	Wait(rw io.ReadWriter, onError chan<- error)
	receive(r io.Reader, onError chan<- error)
	send(w io.Writer, onError chan<- error)
}

// Node represents a Pipe connected to an inside or outside service
type Node struct {
	from, to chan Message // In/Out messages
	err      chan error   // Notify the other side of the channel that an error occured
	addr     string       // Where the Node is connected
}

// OutNode represents a Pipe connected to an outside service
type OutNode Node

func (p OutNode) receive(r io.Reader, onError chan<- error) {
	b := make([]byte, 4096)
	lastMessageEmptyEOF := true
	for {
		n, err := r.Read(b)
		if err != nil {
			select {
			case <-p.err:
				return
			default:
				break
			}
		}
		debugLog.Printf("read %v bytes on %s with err: %v\n", n, p.addr, err)
		if err != nil {
			if err != io.EOF {
				errorLog.Printf("received %v on %s", err, p.addr)
			}
			if !lastMessageEmptyEOF {
				msg := NewMessage([]byte{})
				msg.header.eof = true
				lastMessageEmptyEOF = true
				p.to <- msg
			}
			p.err <- err
			if onError != nil {
				onError <- err
			}
			return
		}
		if n > 0 {
			data := make([]byte, n)
			copy(data, b[:n])
			p.to <- NewMessage(data)
			lastMessageEmptyEOF = false
		}
	}
}

func (p OutNode) send(w io.Writer, onError chan<- error) {
	for {
		select {
		case msg := <-p.from:
			data := msg.payload
			debugLog.Printf("send %v bytes on %s from message with size %v\n", len(data), p.addr, msg.header.size)
			buf := bytes.NewBuffer(data)
			if _, err := io.Copy(w, buf); err != nil {
				errorLog.Printf("could not write %v bytes on %s", len(data), p.addr)
				p.err <- err
				if onError != nil {
					onError <- err
				}
				return
			}
			if msg.EOF() {
				err := io.EOF
				p.err <- err
				if onError != nil {
					onError <- err
				}
				return
			}
		case <-p.err:
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p OutNode) Wait(rw io.ReadWriter, onError chan<- error) {
	go p.receive(rw, onError)
	go p.send(rw, onError)
}

// InNode represents a Pipe connected to another InNode, with custom messages
type InNode Node

func (p InNode) receive(r io.Reader, onError chan<- error) {
	for {
		msg := Message{}
		n, err := msg.ReadFrom(r)
		debugLog.Printf("read message of %v bytes on %s with err: %v\n", n, p.addr, err)
		if err != nil {
			if err != io.EOF {
				errorLog.Printf("received %v on %s", err, p.addr)
			}
			p.err <- err
			if onError != nil {
				onError <- err
			}
			return
		}
		p.to <- msg
	}
}

func (p InNode) send(w io.Writer, onError chan<- error) {
	for {
		select {
		case msg := <-p.from:
			debugLog.Printf("send message %v bytes on %s\n", msg.Size(), p.addr)
			if n, err := msg.WriteTo(w); err != nil {
				errorLog.Printf("could not write %v bytes on %s", n, p.addr)
				p.from <- msg
				if onError != nil {
					onError <- err
				}
				return
			}
		case <-p.err:
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p InNode) Wait(rw io.ReadWriter, onError chan<- error) {
	go p.receive(rw, onError)
	go p.send(rw, onError)
}
