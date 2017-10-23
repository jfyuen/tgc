package main

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"
)

// TODO: Sleeping/Waiting for tests is ugly, find a nice way to Read/Write at the good time

func TestPipeReceive(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	p := OutNode{from: fromCh, to: toCh, err: make(chan error), addr: "test_conn"}
	errCh := make(chan error)

	var b bytes.Buffer
	go p.receive(&b, errCh)
	hello := []byte("hello")
	_, err := b.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	delay := time.Duration(100) * time.Millisecond
	timer := time.NewTimer(delay)

	select {
	case msg := <-toCh:
		if !bytes.Equal(msg.Payload, hello) {
			t.Errorf("received wrong data: %v vs %v", hello, msg.Payload)
		}
	case <-timer.C:
		t.Errorf("data not received after %v", delay)
	case err := <-errCh:
		t.Errorf("received unexpected error %v", err)
	}
}

func TestPipeSend(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	errCh := make(chan error, 1)
	p := OutNode{from: fromCh, to: toCh, err: make(chan error), addr: "test_conn"}

	b := bytes.Buffer{}
	go p.send(&b, errCh)
	msg := Message{Payload: []byte("hello"), EOF: true}
	fromCh <- msg
	<-p.err
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
	outErrCh := make(chan error)
	pOut := OutNode{from: fromCh, to: toCh, err: make(chan error), addr: "out_conn"}
	outRW := bytes.Buffer{}
	go pOut.send(&outRW, outErrCh)

	inErrCh := make(chan error)
	pIn := InNode{from: toCh, to: fromCh, err: make(chan error), addr: "in_conn"}
	inRW := bytes.Buffer{}
	msg := Message{Payload: []byte("hello"), EOF: true}
	encoder := gob.NewEncoder(&inRW)
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("failed encoding message with %v", err)
	}
	pIn.Wait(&inRW, inErrCh)
	time.Sleep(time.Millisecond * 100)
	b := outRW.Bytes()
	if !bytes.Equal(b, msg.Payload) {
		t.Errorf("sent message %v and received message %v are different", msg.Payload, b)
	}
}
