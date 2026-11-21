package videostream

import (
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
)

func TestReceiveRefusesWhatIsNoMeta(t *testing.T) {
	tests := []struct {
		name string
		send func(*websocket.Conn)
		want string
	}{
		{"socket closes first", func(c *websocket.Conn) { _ = c.Close() }, "read video meta"},
		{"another message", func(c *websocket.Conn) { _ = c.WriteJSON(Message{Type: TypeInputError}) }, "unexpected video message"},
		{"meta without content", func(c *websocket.Conn) { _ = c.WriteJSON(Message{Type: TypeMeta}) }, "unexpected video message"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn := serve(t, func(c *websocket.Conn) {
				tc.send(c)
				_, _, _ = c.ReadMessage()
			})

			_, err := Receive(conn, 2*time.Second)

			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRemoteStreamSkipsTextMessagesAndStopsWhenTheSocketEnds(t *testing.T) {
	conn := serve(t, func(c *websocket.Conn) {
		_ = c.WriteJSON(Message{Type: TypeMeta, Meta: &device.VideoMeta{Codec: "h264"}})
		_ = c.WriteJSON(Message{Type: TypeInputError, Message: "bad key"})
		_ = c.WriteMessage(websocket.TextMessage, []byte("not json"))
		_ = c.WriteMessage(websocket.BinaryMessage, EncodePacket(device.VideoPacket{PTS: 7, Data: []byte{1}}))
		_ = c.Close()
	})
	stream, err := Receive(conn, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	packet, err := stream.ReadPacket()
	if err != nil || packet.PTS != 7 {
		t.Fatalf("packet %+v, err %v", packet, err)
	}
	if _, err := stream.ReadPacket(); err == nil {
		t.Fatal("a closed socket must end the stream")
	}
}
