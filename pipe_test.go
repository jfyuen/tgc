package main

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestPipeReceive(t *testing.T) {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	p := InitPipe(fromCh, toCh, "test_conn")

	var b bytes.Buffer
	go p.receive(&b)
	hello := "hello"
	b.Write([]byte(hello))
	delay := .1
	timer := time.NewTimer(time.Millisecond * time.Duration(500))

	select {
	case e := <-toCh:
		s := string(e)
		if s != hello {
			t.Errorf("received wrong data: %v vs %v", hello, s)
		}
	case <-timer.C:
		t.Errorf("data not received after %v ms", delay)
	}
}

func TestSendReceive(t *testing.T) {
	fromCh := make(chan []byte)
	toCh := make(chan []byte)
	p := InitPipe(fromCh, toCh, "test_conn")

	var b bytes.Buffer
	go p.send(&b)
	hello := "hello"
	fromCh <- []byte(hello)
	buf := make([]byte, 10)
	n, err := b.Read(buf)

	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("no data received")
	}
	s := string(buf[:n])
	if s != hello {
		t.Errorf("received wrong data: %v vs %v", hello, s)
	}
}
