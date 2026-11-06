package videostream

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/gorilla/websocket"
)

func TestPacketRoundTrip(t *testing.T) {
	for _, p := range []device.VideoPacket{
		{Config: true, Data: []byte{0, 0, 0, 1, 0x67}},
		{KeyFrame: true, PTS: 1 << 40, Data: []byte{0, 0, 0, 1, 0x65}},
		{PTS: 33333, Data: []byte{0, 0, 1, 0x41}},
	} {
		got, err := DecodePacket(EncodePacket(p))
		if err != nil || got.Config != p.Config || got.KeyFrame != p.KeyFrame || got.PTS != p.PTS || !bytes.Equal(got.Data, p.Data) {
			t.Fatalf("round trip of %+v: %+v %v", p, got, err)
		}
	}
	if _, err := DecodePacket([]byte{1, 2}); err == nil {
		t.Fatal("expected an error for a short packet")
	}
}

// serve upgrades and runs fn with the server side of a websocket.
func serve(t *testing.T, fn func(*websocket.Conn)) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		fn(conn)
	}))
	t.Cleanup(srv.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestPumpAndReceive(t *testing.T) {
	meta := device.VideoMeta{Codec: "h264", Width: 576, Height: 1280, ScreenWidth: 1080, ScreenHeight: 2400}
	src := videostreamtest.New(meta,
		device.VideoPacket{Config: true, Data: []byte{0, 0, 0, 1, 0x67, 0x42}},
		device.VideoPacket{KeyFrame: true, PTS: 10, Data: []byte{0, 0, 0, 1, 0x65}},
	)
	pumped := make(chan error, 1)
	conn := serve(t, func(c *websocket.Conn) { pumped <- Pump(context.Background(), c, src, InjectInto(src)) })

	stream, err := Receive(conn, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if stream.Meta() != meta {
		t.Fatalf("meta %+v", stream.Meta())
	}
	p, err := stream.ReadPacket()
	if err != nil || !p.Config || len(p.Data) != 6 {
		t.Fatalf("first packet %+v %v", p, err)
	}
	p, err = stream.ReadPacket()
	if err != nil || !p.KeyFrame || p.PTS != 10 {
		t.Fatalf("second packet %+v %v", p, err)
	}

	// input goes back to the source stream
	touch := device.InputEvent{Kind: device.InputTouch, Action: device.InputDown, PointerID: 1, X: 10, Y: 20, Width: 576, Height: 1280}
	if err := stream.(device.VideoInput).InjectInput(touch); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-src.Inputs:
		if got != touch {
			t.Fatalf("input %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("input not delivered")
	}

	// hanging up stops the pump and closes the source stream
	_ = stream.Close()
	select {
	case <-src.Closed():
	case <-time.After(2 * time.Second):
		t.Fatal("source stream not closed after the client hung up")
	}
	select {
	case <-pumped:
	case <-time.After(2 * time.Second):
		t.Fatal("pump did not return")
	}
}

func TestReceiveError(t *testing.T) {
	conn := serve(t, func(c *websocket.Conn) {
		_ = WriteError(c, device.ErrUnsupported)
		_, _, _ = c.ReadMessage()
	})
	if _, err := Receive(conn, 2*time.Second); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}

	conn = serve(t, func(c *websocket.Conn) {
		_ = WriteError(c, errors.New("scrcpy failed"))
		_, _, _ = c.ReadMessage()
	})
	if _, err := Receive(conn, 2*time.Second); err == nil || err.Error() != "scrcpy failed" {
		t.Fatalf("expected the node error, got %v", err)
	}
}

func TestRendezvous(t *testing.T) {
	r := NewRendezvous()
	if r.Deliver("unknown", nil) {
		t.Fatal("delivered without a waiter")
	}

	ch, done := r.Expect("s1")
	src := videostreamtest.New(device.VideoMeta{Codec: "h264"})
	conn := serve(t, func(c *websocket.Conn) { _ = Pump(context.Background(), c, src, InjectInto(src)) })
	if !r.Deliver("s1", conn) {
		t.Fatal("not delivered")
	}
	done()
	stream, err := Await(context.Background(), ch, 2*time.Second)
	if err != nil || stream.Meta().Codec != "h264" {
		t.Fatalf("await: %v", err)
	}
	_ = stream.Close()
	if r.Deliver("s1", nil) {
		t.Fatal("id was delivered twice")
	}

	ch, done = r.Expect("s2")
	done()
	if _, err := Await(context.Background(), ch, 50*time.Millisecond); err == nil {
		t.Fatal("expected a timeout")
	}
}
