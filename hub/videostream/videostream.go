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
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

// Pump writes the meta and then every packet of the stream to the websocket
// until the stream ends, a write fails or ctx is done. It closes the stream
// but not the socket. A reader goroutine passes the client's text messages to
// handle (which may be nil) and notices close frames and disconnects; those
// end the pump as well.
func Pump(ctx context.Context, conn *websocket.Conn, stream device.VideoStream, handle InputHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var writeMu sync.Mutex
	reply := func(m Message) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = writeJSON(conn, m)
	}

	go func() {
		defer cancel()
		for {
			kind, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if kind != websocket.TextMessage || handle == nil {
				continue
			}
			var m Message
			if err := json.Unmarshal(data, &m); err != nil {
				reply(Message{Type: TypeInputError, Message: "invalid message: " + err.Error()})
				continue
			}
			handle(m, reply)
		}
	}()
	// a blocked ReadPacket only returns once the stream is closed
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()

	meta := stream.Meta()
	writeMu.Lock()
	err := writeJSON(conn, Message{Type: TypeMeta, Meta: &meta})
	writeMu.Unlock()
	if err != nil {
		return err
	}
	for {
		p, err := stream.ReadPacket()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		writeMu.Lock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		err = conn.WriteMessage(websocket.BinaryMessage, EncodePacket(p))
		writeMu.Unlock()
		if err != nil {
			return err
		}
	}
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

// remoteStream reads a stream that a node pumps into a websocket. Input is
// written back to the node over the same socket.
type remoteStream struct {
	conn      *websocket.Conn
	meta      device.VideoMeta
	writeMu   sync.Mutex
	closeOnce sync.Once
}

// InjectInput implements device.VideoInput by forwarding to the node. Errors
// of the node arrive asynchronously and are only logged there.
func (s *remoteStream) InjectInput(ev device.InputEvent) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return writeJSON(s.conn, Message{Type: TypeInput, Input: &ev})
}

// Receive reads the meta message from a pumped websocket and returns the
// stream. Closing the stream closes the socket.
func Receive(conn *websocket.Conn, timeout time.Duration) (device.VideoStream, error) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	var m Message
	if err := conn.ReadJSON(&m); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read video meta: %w", err)
	}
	_ = conn.SetReadDeadline(time.Time{})
	switch {
	case m.Type == TypeError && m.Unsupported:
		_ = conn.Close()
		return nil, device.ErrUnsupported
	case m.Type == TypeError:
		_ = conn.Close()
		return nil, errors.New(m.Message)
	case m.Type != TypeMeta || m.Meta == nil:
		_ = conn.Close()
		return nil, fmt.Errorf("unexpected video message %q", m.Type)
	}
	return &remoteStream{conn: conn, meta: *m.Meta}, nil
}

func (s *remoteStream) Meta() device.VideoMeta { return s.meta }

func (s *remoteStream) ReadPacket() (device.VideoPacket, error) {
	for {
		kind, data, err := s.conn.ReadMessage()
		if err != nil {
			return device.VideoPacket{}, err
		}
		if kind == websocket.BinaryMessage {
			return DecodePacket(data)
		}
		var m Message
		if json.Unmarshal(data, &m) == nil && m.Type == TypeInputError {
			logrus.Debugf("node rejected video input: %s", m.Message)
		}
	}
}

func (s *remoteStream) Close() error {
	s.closeOnce.Do(func() {
		_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		_ = s.conn.Close()
	})
	return nil
}
