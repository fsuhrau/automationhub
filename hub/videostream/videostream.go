// Package videostream moves device video streams over websockets: from a node
// to the master, and from the master to the browser.
//
// Both hops use the same messages. A text message carries JSON (Message);
// a binary message carries one packet: 1 byte flags (bit 0 config, bit 1 key
// frame), 8 bytes big endian PTS in microseconds, then the packet data.
//
// The client may send "input" messages back (device.InputEvent). The browser
// adds its lease token; the master checks it and forwards the event to the
// node without it.
package videostream

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
)

const (
	flagConfig   = 1
	flagKeyFrame = 2
	headerSize   = 9

	writeTimeout = 10 * time.Second
)

// Message types of text messages.
const (
	TypeMeta  = "meta"
	TypeError = "error"
	// TypeInput is sent by the client to inject input.
	TypeInput = "input"
	// TypeInputError answers a rejected input message; the stream goes on.
	TypeInputError = "inputError"
)

// Message is a JSON control message of a stream.
type Message struct {
	Type string            `json:"type"`
	Meta *device.VideoMeta `json:"meta,omitempty"`
	// Error messages: Unsupported is set when the device can't stream video,
	// so clients fall back to screenshots.
	Message     string `json:"message,omitempty"`
	Unsupported bool   `json:"unsupported,omitempty"`
	// Input messages
	Input *device.InputEvent `json:"input,omitempty"`
	Lease string             `json:"lease,omitempty"`
}

// InputHandler handles a text message from the client. reply sends a
// message back; it is safe to call while packets are written.
type InputHandler func(m Message, reply func(Message))

// EncodePacket encodes a packet as a binary message payload.
func EncodePacket(p device.VideoPacket) []byte {
	buf := make([]byte, headerSize+len(p.Data))
	if p.Config {
		buf[0] |= flagConfig
	}
	if p.KeyFrame {
		buf[0] |= flagKeyFrame
	}
	binary.BigEndian.PutUint64(buf[1:headerSize], uint64(p.PTS))
	copy(buf[headerSize:], p.Data)
	return buf
}

// DecodePacket decodes a binary message payload.
func DecodePacket(data []byte) (device.VideoPacket, error) {
	if len(data) < headerSize {
		return device.VideoPacket{}, fmt.Errorf("video packet too short (%d bytes)", len(data))
	}
	return device.VideoPacket{
		Config:   data[0]&flagConfig != 0,
		KeyFrame: data[0]&flagKeyFrame != 0,
		PTS:      int64(binary.BigEndian.Uint64(data[1:headerSize])),
		Data:     data[headerSize:],
	}, nil
}

// WriteError sends an error message to the client.
func WriteError(conn *websocket.Conn, err error) error {
	return writeJSON(conn, Message{Type: TypeError, Message: err.Error(), Unsupported: errors.Is(err, device.ErrUnsupported)})
}

func writeJSON(conn *websocket.Conn, m Message) error {
	_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return conn.WriteJSON(m)
}

// InjectInto returns a handler that injects input messages into a local
// stream. It doesn't check leases; the master did that already.
func InjectInto(stream device.VideoStream) InputHandler {
	return func(m Message, reply func(Message)) {
		if m.Type != TypeInput || m.Input == nil {
			return
		}
		input, ok := stream.(device.VideoInput)
		if !ok {
			reply(Message{Type: TypeInputError, Message: device.ErrUnsupported.Error(), Unsupported: true})
			return
		}
		if err := input.InjectInput(*m.Input); err != nil {
			reply(Message{Type: TypeInputError, Message: err.Error()})
		}
	}
}
