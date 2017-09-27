package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// MessageHeader contains the size of the payload and if the message contains an EOF
type MessageHeader struct {
	size int16
	eof  bool
}

// Size returns the total size of the header
func (mh MessageHeader) Size() int {
	return binary.Size(mh.size) + binary.Size(mh.eof)
}

// WriteTo writes the MessageHeader into a writer
func (mh MessageHeader) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.LittleEndian, mh.size)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.LittleEndian, mh.eof)
	if err != nil {
		return int64(binary.Size(mh.size)), err
	}
	return int64(mh.Size()), nil
}

// ReadFrom reads from a reader into a MessageHeader
func (mh *MessageHeader) ReadFrom(r io.Reader) (int64, error) {
	err := binary.Read(r, binary.LittleEndian, &mh.size)
	if err != nil {
		return 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &mh.eof); err != nil {
		return int64(binary.Size(mh.size)), err
	}
	return int64(mh.Size()), err
}

// Message is a custom protocol with a small header and the payload
type Message struct {
	header  MessageHeader
	payload []byte
}

// EOF contains an EOF information, to be forwarded to Out nodes
func (m Message) EOF() bool {
	return m.header.eof
}

// WriteTo writes current message information to writer
func (m Message) WriteTo(w io.Writer) (int64, error) {
	n1, err := m.header.WriteTo(w)
	if err != nil {
		return n1, err
	}
	n2, err := w.Write(m.payload)
	if err != nil {
		return n1 + int64(n2), err
	}
	return n1 + int64(n2), nil
}

// ReadFrom reads from a reader and puts data into the Message
func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	n1, err := m.header.ReadFrom(r)
	if err != nil {
		return n1, err
	}
	m.payload = make([]byte, m.header.size)
	n2, err := io.ReadFull(r, m.payload)
	return n1 + int64(n2), err
}

// Size is the total message size: header + payload
func (m Message) Size() int {
	return m.header.Size() + len(m.payload)
}

// NewMessage creates a new message from a payload
func NewMessage(payload []byte) Message {
	if len(payload) > math.MaxInt16 {
		panic(fmt.Sprintf("trying to pass a message of size %v larger than %v", len(payload), math.MaxInt16))
	}
	header := MessageHeader{size: int16(len(payload)), eof: false}
	return Message{header: header, payload: payload}
}
