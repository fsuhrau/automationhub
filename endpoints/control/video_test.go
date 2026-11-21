package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// videoDevice serves a prepared stream and records the options it was asked for.
type videoDevice struct {
	*testkit.FakeDevice
	stream *videostreamtest.Stream
}

func (d videoDevice) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	_, _ = d.FakeDevice.StartVideo(ctx, opts)
	return d.stream, nil
}

func dialVideo(t *testing.T, srv *httptest.Server, path string, header http.Header) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+path, header)
	if conn != nil {
		t.Cleanup(func() { _ = conn.Close() })
	}
	return conn, resp, err
}

func TestVideo(t *testing.T) {
	e := setup(t)
	meta := device.VideoMeta{Codec: "h264", Width: 576, Height: 1280, ScreenWidth: 1080, ScreenHeight: 2400}
	stream := videostreamtest.New(meta, device.VideoPacket{Config: true, Data: []byte{0, 0, 0, 1, 0x67}})
	fake := testkit.NewFakeDevice("video-3")
	e.devices["3"] = videoDevice{FakeDevice: fake, stream: stream}
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	conn, _, err := dialVideo(t, srv, "/api/p/device/3/video?maxSize=99999&maxFps=24", nil)
	if err != nil {
		t.Fatal(err)
	}
	var m videostream.Message
	if err := conn.ReadJSON(&m); err != nil || m.Type != videostream.TypeMeta || *m.Meta != meta {
		t.Fatalf("meta message %+v %v", m, err)
	}
	kind, data, err := conn.ReadMessage()
	if err != nil || kind != websocket.BinaryMessage {
		t.Fatalf("packet: %v", err)
	}
	if p, _ := videostream.DecodePacket(data); !p.Config {
		t.Fatalf("expected the config packet, got %+v", p)
	}
	calls := fake.CallsTo("StartVideo")
	if len(calls) != 1 {
		t.Fatalf("StartVideo calls: %+v", calls)
	}
	if opts := calls[0].Args[0].(device.VideoOptions); opts.MaxSize != maxVideoMaxSize || opts.MaxFPS != 24 || opts.BitRate != defaultVideoBitRate {
		t.Fatalf("options not clamped or defaulted: %+v", opts)
	}

	_ = conn.Close()
	select {
	case <-stream.Closed():
	case <-time.After(2 * time.Second):
		t.Fatal("stream not closed after the viewer left")
	}
}

func TestVideoInputNeedsLease(t *testing.T) {
	e := setup(t)
	stream := videostreamtest.New(device.VideoMeta{Codec: "h264", Width: 576, Height: 1280})
	e.devices["3"] = videoDevice{FakeDevice: testkit.NewFakeDevice("video-3"), stream: stream}
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	conn, _, err := dialVideo(t, srv, "/api/p/device/3/video", nil)
	if err != nil {
		t.Fatal(err)
	}
	var m videostream.Message
	if err := conn.ReadJSON(&m); err != nil || m.Type != videostream.TypeMeta {
		t.Fatalf("meta: %+v %v", m, err)
	}
	touch := device.InputEvent{Kind: device.InputTouch, Action: device.InputDown, X: 1, Y: 2, Width: 576, Height: 1280}

	// without a lease the input is rejected and the stream goes on
	if err := conn.WriteJSON(videostream.Message{Type: videostream.TypeInput, Input: &touch}); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	m = videostream.Message{}
	if err := conn.ReadJSON(&m); err != nil || m.Type != videostream.TypeInputError || !strings.Contains(m.Message, "lock") {
		t.Fatalf("expected a lease error, got %+v %v", m, err)
	}

	token := e.lock(t, "3", "alice")
	if err := conn.WriteJSON(videostream.Message{Type: videostream.TypeInput, Input: &touch, Lease: token}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-stream.Inputs:
		if got != touch {
			t.Fatalf("input %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("input not injected")
	}
	select {
	case <-stream.Inputs:
		t.Fatal("the input without lease was injected")
	default:
	}
}

func TestVideoIgnoresOtherMessages(t *testing.T) {
	h := (&Handler{}).videoInput(nil, nil)
	h(videostream.Message{Type: videostream.TypeMeta}, func(videostream.Message) { t.Fatal("unexpected reply") })
	h(videostream.Message{Type: videostream.TypeInput}, func(videostream.Message) { t.Fatal("unexpected reply") })
}

// plainStream hides the input capability of the stream it wraps.
type plainStream struct{ device.VideoStream }

func TestVideoInputUnsupported(t *testing.T) {
	svc := hubcontrol.New(nil)
	dev := plainDevice{testkit.NewFakeDevice("plain")}
	token := ""
	if l, err := svc.Acquire(dev, "alice"); err == nil {
		token = l.Token
	}
	var reply videostream.Message
	h := (&Handler{svc: svc}).videoInput(dev, plainStream{videostreamtest.New(device.VideoMeta{})})
	h(videostream.Message{Type: videostream.TypeInput, Lease: token, Input: &device.InputEvent{}}, func(m videostream.Message) { reply = m })
	if reply.Type != videostream.TypeInputError || !reply.Unsupported {
		t.Fatalf("expected an unsupported input error, got %+v", reply)
	}
}

func TestVideoUnsupported(t *testing.T) {
	e := setup(t)
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	conn, _, err := dialVideo(t, srv, "/api/p/device/2/video", nil)
	if err != nil {
		t.Fatal(err)
	}
	var m videostream.Message
	if err := conn.ReadJSON(&m); err != nil || m.Type != videostream.TypeError || !m.Unsupported {
		t.Fatalf("expected an unsupported error, got %+v %v", m, err)
	}
}

func TestVideoRejectsCrossOrigin(t *testing.T) {
	e := setup(t)
	srv := httptest.NewServer(e.router)
	defer srv.Close()

	_, resp, err := dialVideo(t, srv, "/api/p/device/1/video", http.Header{"Origin": {"https://evil.example"}})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a foreign origin, got %v", err)
	}
	if _, _, err := dialVideo(t, srv, "/api/p/device/1/video", http.Header{"Origin": {srv.URL}}); err != nil {
		t.Fatalf("same origin rejected: %v", err)
	}
}

func TestCheckOrigin(t *testing.T) {
	h := (&Handler{}).WithAllowedOrigins([]string{"https://hub.example/", "https://*.skip"})
	tests := []struct {
		name, origin string
		want         bool
	}{
		{"no origin", "", true},
		{"same host", "http://app.local", true},
		{"allowed with trailing slash", "https://hub.example", true},
		{"foreign", "https://evil.example", false},
		{"unparsable", "http://%zz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://app.local/video", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := h.checkOrigin(r); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
	any := (&Handler{}).WithAllowedOrigins([]string{"*"})
	r := httptest.NewRequest(http.MethodGet, "http://app.local/video", nil)
	r.Header.Set("Origin", "https://anywhere.example")
	if !any.checkOrigin(r) {
		t.Fatal("wildcard origin rejected")
	}
}

func TestVideoOptions(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?maxSize=-5&maxFps=0&bitRate=1", nil)
	got := videoOptions(c)
	want := device.VideoOptions{MaxSize: 0, MaxFPS: 1, BitRate: minVideoBitRate}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPumpVideoLogsAnEndedStream(t *testing.T) {
	stream := videostreamtest.New(device.VideoMeta{Codec: "h264"})
	dev := videoDevice{FakeDevice: testkit.NewFakeDevice("video-3"), stream: stream}
	h := &Handler{svc: hubcontrol.New(nil)}
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.Close() // the viewer is gone: writing the meta message fails
		h.pumpVideo(conn, dev, device.VideoOptions{})
	}))
	defer srv.Close()
	if _, _, err := dialVideo(t, srv, "/", nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pumpVideo did not return")
	}
}
