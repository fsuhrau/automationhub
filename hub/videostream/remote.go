package videostream

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

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
