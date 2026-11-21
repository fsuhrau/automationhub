package videostream

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/gorilla/websocket"
)

// failingStream returns its meta and then an error instead of packets.
type failingStream struct {
	device.VideoStream
	err error
}

func (s *failingStream) ReadPacket() (device.VideoPacket, error) { return device.VideoPacket{}, s.err }

func TestPumpReturnsTheErrorOfAStreamThatFails(t *testing.T) {
	boom := errors.New("scrcpy died")
	pumped := make(chan error, 1)
	conn := serve(t, func(c *websocket.Conn) {
		pumped <- Pump(context.Background(), c, Feed{Stream: &failingStream{VideoStream: videostreamtest.New(device.VideoMeta{}), err: boom}})
	})
	defer func() { _ = conn.Close() }()

	select {
	case err := <-pumped:
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump did not return")
	}
}

func TestPumpReturnsTheErrorOfAFailingMetaWrite(t *testing.T) {
	pumped := make(chan error, 1)
	conn := serve(t, func(c *websocket.Conn) {
		_ = c.Close()
		pumped <- Pump(context.Background(), c, Feed{Stream: videostreamtest.New(device.VideoMeta{})})
	})
	defer func() { _ = conn.Close() }()

	select {
	case err := <-pumped:
		if err == nil {
			t.Fatal("no error for a closed socket")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump did not return")
	}
}

// pumpWith starts a pump on a source with the handler and returns the client
// side of the socket after the meta message was read.
func pumpWith(t *testing.T, handle InputHandler) *websocket.Conn {
	t.Helper()
	src := videostreamtest.New(device.VideoMeta{Codec: "h264"})
	conn := serve(t, func(c *websocket.Conn) { _ = Pump(context.Background(), c, Feed{Stream: src, Input: handle}) })
	var meta Message
	if err := conn.ReadJSON(&meta); err != nil || meta.Type != TypeMeta {
		t.Fatalf("meta %+v, err %v", meta, err)
	}
	return conn
}

func TestPumpAnswersAMessageThatIsNoJSON(t *testing.T) {
	conn := pumpWith(t, InjectInto(videostreamtest.New(device.VideoMeta{})))

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{")); err != nil {
		t.Fatal(err)
	}

	var reply Message
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&reply); err != nil || reply.Type != TypeInputError || !strings.Contains(reply.Message, "invalid message") {
		t.Fatalf("reply %+v, err %v", reply, err)
	}
}

func TestPumpIgnoresClientMessagesWithoutAHandler(t *testing.T) {
	conn := pumpWith(t, nil)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"input"}`)); err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte{1}); err != nil {
		t.Fatal(err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("the pump answered a message it has no handler for")
	}
}

func TestInjectIntoAnswersWhatItCannotInject(t *testing.T) {
	tests := []struct {
		name        string
		stream      device.VideoStream
		msg         Message
		wantReply   bool
		unsupported bool
	}{
		{"not an input message", videostreamtest.New(device.VideoMeta{}), Message{Type: TypeMeta}, false, false},
		{"input without event", videostreamtest.New(device.VideoMeta{}), Message{Type: TypeInput}, false, false},
		{"stream without input", noInputStream{}, Message{Type: TypeInput, Input: &device.InputEvent{}}, true, true},
		{"input fails", failingInput{}, Message{Type: TypeInput, Input: &device.InputEvent{}}, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var replies []Message

			InjectInto(tc.stream)(tc.msg, func(m Message) { replies = append(replies, m) })

			if (len(replies) == 1) != tc.wantReply || (tc.wantReply && replies[0].Unsupported != tc.unsupported) {
				t.Fatalf("replies %+v", replies)
			}
		})
	}
}

type noInputStream struct{}

func (noInputStream) Meta() device.VideoMeta                  { return device.VideoMeta{} }
func (noInputStream) ReadPacket() (device.VideoPacket, error) { return device.VideoPacket{}, nil }
func (noInputStream) Close() error                            { return nil }

type failingInput struct{ noInputStream }

func (failingInput) InjectInput(device.InputEvent) error { return errors.New("rejected") }

// closingStream breaks the socket while it hands out its first packet.
type closingStream struct {
	device.VideoStream
	conn *websocket.Conn
}

func (s *closingStream) ReadPacket() (device.VideoPacket, error) {
	_ = s.conn.UnderlyingConn().Close()
	return device.VideoPacket{Data: []byte{1}}, nil
}

func TestPumpReturnsTheErrorOfAFailingPacketWrite(t *testing.T) {
	pumped := make(chan error, 1)
	conn := serve(t, func(c *websocket.Conn) {
		pumped <- Pump(context.Background(), c, Feed{Stream: &closingStream{VideoStream: videostreamtest.New(device.VideoMeta{}), conn: c}})
	})
	defer func() { _ = conn.Close() }()

	select {
	case err := <-pumped:
		if err == nil {
			t.Fatal("no error for a broken socket")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump did not return")
	}
}
