package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestMessageWriteTo(t *testing.T) {
	data := []byte("hello")
	msg := NewMessage(data)
	var buf bytes.Buffer
	n, err := msg.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(msg.Size()) {
		t.Errorf("wrote %v bytes, expected %v\n", n, len(data))
	}
	b := buf.Bytes()
	if len(b) != msg.Size() {
		t.Errorf("read buffer of size %v bytes does not match message of size %v bytes", len(b), msg.Size())
	}
	var size int16
	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &size)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	if int16(len(data)) != size {
		t.Errorf("expected size %v, got %v", len(data), size)
	}

	payload := b[len(b)-len(data):]
	if string(data) != string(payload) {
		t.Errorf("expected []bytes %v, got %v", data, payload)
	}
}

func TestMessageReadFrom(t *testing.T) {
	data := []byte{5, 0, 0, 104, 101, 108, 108, 111}
	r := bytes.NewReader(data)
	msg := Message{}
	msg.ReadFrom(r)
	if string(data[int16(len(data))-msg.header.size:]) != string(msg.payload) {
		t.Errorf("expected []bytes %v, got %v", data, msg.payload)
	}
}
