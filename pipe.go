package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
)

// Pipe reads and writes on a connection, results are sent over channels
type Pipe interface {
	Wait(ctx context.Context, rw io.ReadWriter)
	receive(ctx context.Context, r io.Reader)
	send(ctx context.Context, w io.Writer)
}

// Node represents a Pipe connected to an inside or outside service
type Node struct {
	from, to chan Message       // In/Out messages
	addr     string             // Where the Node is connected
	cancel   context.CancelFunc // Cancel the connections, linked to the Context
}

// OutNode represents a Pipe connected to an outside service
type OutNode struct {
	Node
	firstMessage *Message
}

func (p OutNode) receive(ctx context.Context, r io.Reader) {
	b := make([]byte, 4096)
	for {
		n, err := r.Read(b)
		if n == 0 && ctx.Err() != nil {
			return
		}
		data := make([]byte, n)
		copy(data, b[:n])
		p.to <- Message{Payload: data, EOF: err == io.EOF}
		debugLog.Printf("read %v bytes on %s with err: %v\n", n, p.addr, err)
		if err != nil {
			if err != io.EOF && ctx.Err() == nil {
				errorLog.Printf("received %v on %s", err, p.addr)
			}
			p.cancel()
			return
		}
	}
}

func (p OutNode) sendData(ctx context.Context, w io.Writer, msg Message) error {
	data := msg.Payload
	debugLog.Printf("send %v bytes on %s\n", len(data), p.addr)
	buf := bytes.NewBuffer(data)
	if _, err := io.Copy(w, buf); err != nil {
		errorLog.Printf("could not write %v bytes on %s", len(data), p.addr)
		p.cancel()
		return err
	}
	if msg.EOF {
		p.cancel()
		return io.EOF
	}
	return nil
}

func (p OutNode) send(ctx context.Context, w io.Writer) {
	if p.firstMessage != nil {
		if err := p.sendData(ctx, w, *p.firstMessage); err != nil {
			return
		}
	}
	for {
		select {
		case msg := <-p.from:
			if err := p.sendData(ctx, w, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p OutNode) Wait(ctx context.Context, rw io.ReadWriter) {
	go p.receive(ctx, rw)
	go p.send(ctx, rw)
}

// InNode represents a Pipe connected to another InNode, with custom messages
type InNode Node

func (p InNode) receive(ctx context.Context, r io.Reader) {
	decoder := gob.NewDecoder(r)
	for {
		msg := Message{}
		err := decoder.Decode(&msg)
		debugLog.Printf("decode message of %v bytes on %s with err: %v\n", msg.Size(), p.addr, err)
		if err != nil {
			if err != io.EOF {
				errorLog.Printf("received %v on %s", err, p.addr)
			}
			p.cancel()
			return
		}
		p.to <- msg
	}
}

func (p InNode) send(ctx context.Context, w io.Writer) {
	encoder := gob.NewEncoder(w)
	for {
		select {
		case msg := <-p.from:
			debugLog.Printf("send message %v bytes on %s\n", msg.Size(), p.addr)
			if err := encoder.Encode(msg); err != nil {
				errorLog.Printf("could not encode message on %s with error %v", p.addr, err)
				p.from <- msg
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Wait will read data on connection and write back results
func (p InNode) Wait(ctx context.Context, rw io.ReadWriter) {
	go p.receive(ctx, rw)
	go p.send(ctx, rw)
}
