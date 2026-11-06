package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type fakeDevice struct {
	device.Device
	id string

	mu     sync.Mutex
	locked bool
	taps   [][2]int
	keys   []string
	png    []byte

	videoOpts device.VideoOptions
}

func (d *fakeDevice) DeviceID() string     { return d.id }
func (d *fakeDevice) IsAppConnected() bool { return false }
func (d *fakeDevice) Lock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.locked {
		return fmt.Errorf("locked")
	}
	d.locked = true
	return nil
}
func (d *fakeDevice) Unlock() error  { return d.Release() }
func (d *fakeDevice) Release() error { d.mu.Lock(); d.locked = false; d.mu.Unlock(); return nil }
func (d *fakeDevice) IsLocked() bool { d.mu.Lock(); defer d.mu.Unlock(); return d.locked }
func (d *fakeDevice) GetScreenshot() ([]byte, int, int, error) {
	return d.png, 4, 2, nil
}

// inputDevice additionally supports native input.
type inputDevice struct{ *fakeDevice }

func (d inputDevice) Tap(x, y int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.taps = append(d.taps, [2]int{x, y})
	return nil
}
func (d inputDevice) Swipe(int, int, int, int, time.Duration) error { return nil }
func (d inputDevice) InputText(string) error                        { return nil }
func (d inputDevice) PressKey(k string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.keys = append(d.keys, k)
	return nil
}

func testPNG(t *testing.T) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setup(t *testing.T) (*gin.Engine, map[string]device.Device) {
	gin.SetMode(gin.TestMode)
	png := testPNG(t)
	devices := map[string]device.Device{
		"1": inputDevice{&fakeDevice{id: "android-1", png: png}},
		"2": &fakeDevice{id: "plain-2", png: png}, // no native input
	}
	resolve := func(c *gin.Context) (device.Device, bool) {
		d, ok := devices[c.Param("device_id")]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
		}
		return d, ok
	}
	resolveApp := func(c *gin.Context, id uint) (*app.Parameter, error) {
		return &app.Parameter{Identifier: "com.example"}, nil
	}
	owner := func(c *gin.Context) string { return c.GetHeader("X-Test-User") }

	r := gin.New()
	NewHandler(hubcontrol.New(nil), resolve, resolveApp, owner, "localhost").Register(r.Group("/api/p"))
	return r, devices
}

func do(r http.Handler, method, path, user, leaseToken, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", user)
	if leaseToken != "" {
		req.Header.Set(LeaseHeader, leaseToken)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestLeaseFlow(t *testing.T) {
	r, devices := setup(t)
	dev := devices["1"].(inputDevice)

	if w := do(r, "POST", "/api/p/device/1/tap", "alice", "", `{"x":1,"y":2}`); w.Code != http.StatusPreconditionRequired {
		t.Fatalf("tap without lease: expected 428, got %d %s", w.Code, w.Body)
	}

	w := do(r, "POST", "/api/p/device/1/lock", "alice", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("lock: %d %s", w.Code, w.Body)
	}
	var lease hubcontrol.Lease
	_ = json.Unmarshal(w.Body.Bytes(), &lease)
	if lease.Token == "" || lease.Owner != "alice" {
		t.Fatalf("unexpected lease %+v", lease)
	}

	if w := do(r, "POST", "/api/p/device/1/lock", "bob", "", ""); w.Code != http.StatusConflict {
		t.Fatalf("second owner: expected 409, got %d", w.Code)
	}
	if w := do(r, "POST", "/api/p/device/1/tap", "bob", "bogus", `{"x":1,"y":2}`); w.Code != http.StatusConflict {
		t.Fatalf("wrong token: expected 409, got %d", w.Code)
	}
	if w := do(r, "POST", "/api/p/device/1/tap", "alice", lease.Token, `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("tap without coordinates: expected 400, got %d", w.Code)
	}
	if w := do(r, "POST", "/api/p/device/1/tap", "alice", lease.Token, `{"x":10,"y":20}`); w.Code != http.StatusNoContent {
		t.Fatalf("tap: expected 204, got %d %s", w.Code, w.Body)
	}
	if w := do(r, "POST", "/api/p/device/1/key", "alice", lease.Token, `{"key":"back"}`); w.Code != http.StatusNoContent {
		t.Fatalf("key: expected 204, got %d", w.Code)
	}
	if len(dev.taps) != 1 || dev.taps[0] != [2]int{10, 20} || len(dev.keys) != 1 {
		t.Fatalf("input not forwarded: taps=%v keys=%v", dev.taps, dev.keys)
	}

	w = do(r, "GET", "/api/p/device/1/lock", "bob", "", "")
	var status lockResponse
	_ = json.Unmarshal(w.Body.Bytes(), &status)
	if !status.Locked || status.Lease == nil || status.Lease.Owner != "alice" || status.Lease.Token != "" {
		t.Fatalf("unexpected lock status %s", w.Body)
	}

	if w := do(r, "DELETE", "/api/p/device/1/lock", "alice", lease.Token, ""); w.Code != http.StatusNoContent {
		t.Fatalf("release: expected 204, got %d", w.Code)
	}
	if dev.IsLocked() {
		t.Fatal("device still locked after release")
	}
}

func TestUnsupportedInput(t *testing.T) {
	r, _ := setup(t)
	w := do(r, "POST", "/api/p/device/2/lock", "alice", "", "")
	var lease hubcontrol.Lease
	_ = json.Unmarshal(w.Body.Bytes(), &lease)
	if w := do(r, "POST", "/api/p/device/2/tap", "alice", lease.Token, `{"x":1,"y":1}`); w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
	if w := do(r, "GET", "/api/p/device/2/logs", "alice", "", ""); w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for logs, got %d", w.Code)
	}
}

func TestLockedByTestRun(t *testing.T) {
	r, devices := setup(t)
	_ = devices["1"].Lock() // a test run holds it
	if w := do(r, "POST", "/api/p/device/1/lock", "alice", "", ""); w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	w := do(r, "GET", "/api/p/device/1/lock", "alice", "", "")
	var status lockResponse
	_ = json.Unmarshal(w.Body.Bytes(), &status)
	if !status.LockedByTest || status.Lease != nil {
		t.Fatalf("expected lockedByTest, got %s", w.Body)
	}
}

func TestScreenshot(t *testing.T) {
	r, _ := setup(t)
	w := do(r, "GET", "/api/p/device/1/screenshot", "alice", "", "")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("unexpected response %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	if w.Header().Get("X-Screen-Width") != "4" || w.Header().Get("X-Screen-Height") != "2" {
		t.Fatal("missing screen size headers")
	}
}

func TestAppNotConnected(t *testing.T) {
	r, _ := setup(t)
	if w := do(r, "GET", "/api/p/device/1/hierarchy", "alice", "", ""); w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "no app connected") {
		t.Fatalf("expected app not connected error, got %d %s", w.Code, w.Body)
	}
}

func TestStreamSendsJPEGFrames(t *testing.T) {
	r, _ := setup(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+"/api/p/device/1/stream?fps=5", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "multipart/x-mixed-replace") {
		t.Fatalf("unexpected content type %s", resp.Header.Get("Content-Type"))
	}
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	head := string(buf[:n])
	if !strings.Contains(head, "--frame") || !strings.Contains(head, "image/jpeg") {
		t.Fatalf("no jpeg frame in stream: %q", head)
	}
}

// videoDevice additionally streams video.
type videoDevice struct {
	*fakeDevice
	stream *videostreamtest.Stream
}

func (d videoDevice) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.videoOpts = opts
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
	r, devices := setup(t)
	meta := device.VideoMeta{Codec: "h264", Width: 576, Height: 1280, ScreenWidth: 1080, ScreenHeight: 2400}
	stream := videostreamtest.New(meta, device.VideoPacket{Config: true, Data: []byte{0, 0, 0, 1, 0x67}})
	fake := &fakeDevice{id: "video-3"}
	devices["3"] = videoDevice{fakeDevice: fake, stream: stream}
	srv := httptest.NewServer(r)
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
	fake.mu.Lock()
	opts := fake.videoOpts
	fake.mu.Unlock()
	if opts.MaxSize != maxVideoMaxSize || opts.MaxFPS != 24 || opts.BitRate != defaultVideoBitRate {
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
	r, devices := setup(t)
	stream := videostreamtest.New(device.VideoMeta{Codec: "h264", Width: 576, Height: 1280})
	devices["3"] = videoDevice{fakeDevice: &fakeDevice{id: "video-3"}, stream: stream}
	srv := httptest.NewServer(r)
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

	w := do(r, "POST", "/api/p/device/3/lock", "alice", "", "")
	var lease hubcontrol.Lease
	_ = json.Unmarshal(w.Body.Bytes(), &lease)
	if err := conn.WriteJSON(videostream.Message{Type: videostream.TypeInput, Input: &touch, Lease: lease.Token}); err != nil {
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

func TestVideoUnsupported(t *testing.T) {
	r, _ := setup(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	conn, _, err := dialVideo(t, srv, "/api/p/device/1/video", nil)
	if err != nil {
		t.Fatal(err)
	}
	var m videostream.Message
	if err := conn.ReadJSON(&m); err != nil || m.Type != videostream.TypeError || !m.Unsupported {
		t.Fatalf("expected an unsupported error, got %+v %v", m, err)
	}
}

func TestVideoRejectsCrossOrigin(t *testing.T) {
	r, _ := setup(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	_, resp, err := dialVideo(t, srv, "/api/p/device/1/video", http.Header{"Origin": {"https://evil.example"}})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a foreign origin, got %v", err)
	}
	if _, _, err := dialVideo(t, srv, "/api/p/device/1/video", http.Header{"Origin": {srv.URL}}); err != nil {
		t.Fatalf("same origin rejected: %v", err)
	}
}
