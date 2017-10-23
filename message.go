package main

import (
	"encoding/binary"
)

// Message is a custom protocol with a small header and the payload
type Message struct {
	EOF     bool
	Payload []byte
}

// Size is the total size of the message in bytes, without gob overhead
func (m Message) Size() int {
	return len(m.Payload) + binary.Size(m.EOF)
}
