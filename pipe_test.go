package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"testing"
	"time"
)

func TestPipeReceive(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	delay := time.Duration(100) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	p := OutNode{from: fromCh, to: toCh, cancel: cancel, addr: "test_conn"}

	var b bytes.Buffer
	go p.receive(ctx, &b)
	hello := []byte("hello")
	_, err := b.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-toCh:
		if !bytes.Equal(msg.Payload, hello) {
			t.Errorf("received wrong data: %v vs %v", hello, msg.Payload)
		}
		cancel()
	case <-ctx.Done():
		t.Errorf("data not received after %v", delay)
	}
	if ctx.Err() != context.Canceled {
		t.Errorf("received context error %v", ctx.Err())
	}
}

func TestPipeSend(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	ctx, cancel := context.WithCancel(context.Background())
	p := OutNode{from: fromCh, to: toCh, cancel: cancel, addr: "test_conn"}

	b := bytes.Buffer{}
	go p.send(ctx, &b)
	msg := Message{Payload: []byte("hello"), EOF: true}
	fromCh <- msg
	<-ctx.Done()
	received := b.Bytes()
	if len(received) == 0 {
		t.Fatal("no data received")
	}
	if !bytes.Equal(received, msg.Payload) {
		t.Errorf("received wrong data: %v vs %v", msg.Payload, received)
	}
}

func TestInOutNode(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	outCtx, outCancel := context.WithCancel(context.Background())
	pOut := OutNode{from: fromCh, to: toCh, cancel: outCancel, addr: "out_conn"}
	inRead, inWrite := io.Pipe()   // Inside socket, receiving messages
	outRead, outWrite := io.Pipe() // Outside socket, forwarding clear text
	go pOut.send(outCtx, outWrite)

	inCtx, inCancel := context.WithCancel(context.Background())

	pIn := InNode{from: toCh, to: fromCh, cancel: inCancel, addr: "in_conn"}
	go pIn.receive(inCtx, inRead)
	msg := Message{Payload: []byte("hello"), EOF: true}
	encoder := gob.NewEncoder(inWrite)
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("failed encoding message with %v", err)
	}
	buf := make([]byte, 10, 10)
	n, err := outRead.Read(buf)
	if err != nil {
		t.Fatalf("failed reading from out node with %v", err)
	}
	if !bytes.Equal(buf[:n], msg.Payload) {
		t.Errorf("sent message %v and received message %v are different", msg.Payload, buf)
	}
}
