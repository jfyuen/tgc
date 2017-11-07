package main

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func listenLocal(addr string, listenCh chan []byte, errCh chan error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		errCh <- err
		return
	}
	conn, err := ln.Accept()
	if err != nil {
		errCh <- err
		return
	}
	buf := make([]byte, 10)
	n, err := conn.Read(buf)
	if err != nil {
		errCh <- err
		return
	}
	listenCh <- buf[:n]
}

func TestConnectListen(t *testing.T) {
	listenCh := make(chan []byte)
	listenErrCh := make(chan error)
	connectOutAddr := "127.0.0.1:60003"
	go listenLocal(connectOutAddr, listenCh, listenErrCh)

	listenOutAddr := "127.0.0.1:60001"
	tunnelAddr := "127.0.0.1:60002"
	listener := Listener{from: listenOutAddr, to: tunnelAddr}
	go listener.Listen()

	time.Sleep(50 * time.Millisecond) // Wait for listener to be ready (do it a better way)

	connector := Connector{dst: tunnelAddr, src: connectOutAddr, interval: 1}
	go connector.Connect()

	conn, err := net.Dial("tcp", listenOutAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	sent := []byte("hello")
	_, err = conn.Write(sent)
	if err != nil {
		t.Fatal(err)
	}
	delay := time.Duration(50) * time.Millisecond // test timeout
	timer := time.NewTimer(delay)
	select {
	case err := <-listenErrCh:
		t.Fatal(err)
	case msg := <-listenCh:
		if !bytes.Equal(msg, sent) {
			t.Fatalf("sent (%v) and received (%v) message not equals", sent, msg)
		}
	case <-timer.C:
		t.Fatalf("nothing received after %v", delay)
	}
}
