package main

import (
	"bytes"
	"testing"
	"time"
)

func TestPipeReceive(t *testing.T) {
	fromCh := make(chan Message)
	toCh := make(chan Message)
	p := OutNode{from: fromCh, to: toCh, err: make(chan error), addr: "test_conn"}
	errCh := make(chan error)

	var b bytes.Buffer
	go p.receive(&b, errCh)
	hello := "hello"
	_, err := b.Write([]byte(hello))
	if err != nil {
		t.Fatal(err)
	}

	delay := time.Duration(100) * time.Millisecond
	timer := time.NewTimer(delay)

	select {
	case e := <-toCh:
		s := string(e.payload)
		if s != hello {
			t.Errorf("received wrong data: %v vs %v", hello, s)
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
	errCh := make(chan error)
	p := OutNode{from: fromCh, to: toCh, err: make(chan error), addr: "test_conn"}

	hello := "hello"
	b := bytes.Buffer{}
	go p.send(&b, errCh)
	fromCh <- NewMessage([]byte(hello))
	time.Sleep(time.Millisecond * 100)
	received := b.Bytes()
	if len(received) == 0 {
		t.Fatal("no data received")
	}
	s := string(received)
	if s != hello {
		t.Errorf("received wrong data: %v vs %v", hello, s)
	}
}
