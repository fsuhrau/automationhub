package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeDevice struct {
	device.Device
	id string

	mu     sync.Mutex
	locked bool
	taps   [][2]int
	png    []byte
}

func (d *fakeDevice) DeviceID() string                         { return d.id }
func (d *fakeDevice) DeviceState() device.State                { return device.StateBooted }
func (d *fakeDevice) DeviceParameter() map[string]string       { return nil }
func (d *fakeDevice) IsAppConnected() bool                     { return false }
func (d *fakeDevice) GetScreenshot() ([]byte, int, int, error) { return d.png, 2000, 1000, nil }
func (d *fakeDevice) Lock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.locked {
		return errors.New("locked")
	}
	d.locked = true
	return nil
}
func (d *fakeDevice) Unlock() error  { return d.Release() }
func (d *fakeDevice) Release() error { d.mu.Lock(); d.locked = false; d.mu.Unlock(); return nil }
func (d *fakeDevice) IsLocked() bool { d.mu.Lock(); defer d.mu.Unlock(); return d.locked }
func (d *fakeDevice) Tap(x, y int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.taps = append(d.taps, [2]int{x, y})
	return nil
}
func (d *fakeDevice) Swipe(int, int, int, int, time.Duration) error { return nil }
func (d *fakeDevice) InputText(string) error                        { return nil }
func (d *fakeDevice) PressKey(string) error                         { return nil }

type devicesManager = manager.Devices

type fakeDevices struct {
	devicesManager
	dev *fakeDevice
}

func (m *fakeDevices) GetDevice(id string) (device.Device, string) {
	if id == m.dev.id {
		return m.dev, "fake"
	}
	return nil, ""
}

type fakeBackend struct {
	Backend
}

func (fakeBackend) FindDevice(project *models.Project, id uint) (*models.Device, error) {
	if id != 1 {
		return nil, fmt.Errorf("%w: device %d", errNotFound, id)
	}
	d := &models.Device{DeviceIdentifier: "dev-1", Name: "Pixel", PlatformType: models.PlatformTypeAndroid}
	d.ID = 1
	return d, nil
}

func (b fakeBackend) ProjectDevices(project *models.Project) ([]models.Device, error) {
	d, _ := b.FindDevice(project, 1)
	return []models.Device{*d}, nil
}

func testPNG(t *testing.T, w, h int) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setup(t *testing.T) (*httptest.Server, *fakeDevice) {
	dev := &fakeDevice{id: "dev-1", png: testPNG(t, 2000, 1000)}
	h := NewHandler(Deps{
		Backend: fakeBackend{},
		Devices: &fakeDevices{dev: dev},
		Control: hubcontrol.New(nil),
		Runs:    runs.NewTracker(nil),
	})
	project := &models.Project{Identifier: "p1"}
	project.ID = 1
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r, project, r.Header.Get("X-Caller"))
	}))
	t.Cleanup(srv.Close)
	return srv, dev
}

type callerTransport struct{ caller string }

func (c callerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("X-Caller", c.caller)
	return http.DefaultTransport.RoundTrip(r)
}

func connect(t *testing.T, srv *httptest.Server, caller string) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:   srv.URL,
		HTTPClient: &http.Client{Transport: callerTransport{caller}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return res
}

func text(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestToolsAreListedWithAnnotations(t *testing.T) {
	srv, _ := setup(t)
	cs := connect(t, srv, "alice")
	defer cs.Close()

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*mcp.Tool{}
	for _, tool := range res.Tools {
		byName[tool.Name] = tool
	}
	for _, name := range []string{"list_devices", "run_test", "wait_for_test_run", "get_protocol_logs", "lock_device", "screenshot", "tap", "stop_app"} {
		if byName[name] == nil {
			t.Errorf("tool %s missing", name)
		}
	}
	if !byName["screenshot"].Annotations.ReadOnlyHint {
		t.Error("screenshot should be read-only")
	}
	if d := byName["stop_app"].Annotations.DestructiveHint; d == nil || !*d {
		t.Error("stop_app should be destructive")
	}
}

func TestControlNeedsLockAndLockIsPerSession(t *testing.T) {
	srv, dev := setup(t)
	alice := connect(t, srv, "alice")
	bob := connect(t, srv, "bob")
	defer bob.Close()

	res := call(t, alice, "tap", map[string]any{"deviceId": 1, "x": 5, "y": 6})
	if !res.IsError || !strings.Contains(text(res), "lock_device") {
		t.Fatalf("tap without lock should fail with a hint, got %q", text(res))
	}

	if res := call(t, alice, "lock_device", map[string]any{"deviceId": 1}); res.IsError {
		t.Fatalf("lock failed: %s", text(res))
	}
	if res := call(t, bob, "lock_device", map[string]any{"deviceId": 1}); !res.IsError {
		t.Fatal("second session must not get the lock")
	}
	if res := call(t, bob, "tap", map[string]any{"deviceId": 1, "x": 1, "y": 1}); !res.IsError {
		t.Fatal("other session must not control the device")
	}

	if res := call(t, alice, "tap", map[string]any{"deviceId": 1, "x": 5, "y": 6}); res.IsError {
		t.Fatalf("tap failed: %s", text(res))
	}
	dev.mu.Lock()
	taps := append([][2]int(nil), dev.taps...)
	dev.mu.Unlock()
	if len(taps) != 1 || taps[0] != [2]int{5, 6} {
		t.Fatalf("tap not forwarded: %v", taps)
	}

	// closing the session releases the device
	if err := alice.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for dev.IsLocked() && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if dev.IsLocked() {
		t.Fatal("device still locked after the session closed")
	}
	if res := call(t, bob, "lock_device", map[string]any{"deviceId": 1}); res.IsError {
		t.Fatalf("lock after release failed: %s", text(res))
	}
}

func TestScreenshotIsDownscaled(t *testing.T) {
	srv, _ := setup(t)
	cs := connect(t, srv, "alice")
	defer cs.Close()

	res := call(t, cs, "screenshot", map[string]any{"deviceId": 1})
	if res.IsError {
		t.Fatalf("screenshot failed: %s", text(res))
	}
	var img *mcp.ImageContent
	for _, c := range res.Content {
		if ic, ok := c.(*mcp.ImageContent); ok {
			img = ic
		}
	}
	if img == nil || img.MIMEType != "image/jpeg" {
		t.Fatal("no jpeg image content")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(img.Data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != defaultScreenshotWidth {
		t.Fatalf("expected width %d, got %d", defaultScreenshotWidth, cfg.Width)
	}
	if !strings.Contains(text(res), "2000x1000") {
		t.Fatalf("missing original size in %q", text(res))
	}
}

func TestUnknownDevice(t *testing.T) {
	srv, _ := setup(t)
	cs := connect(t, srv, "alice")
	defer cs.Close()
	res := call(t, cs, "get_device", map[string]any{"deviceId": 42})
	if !res.IsError || !strings.Contains(text(res), "not found") {
		t.Fatalf("expected not found, got %q", text(res))
	}
	res = call(t, cs, "list_devices", map[string]any{"platform": "android"})
	if res.IsError || !strings.Contains(text(res), "Pixel") {
		t.Fatalf("list_devices failed: %s", text(res))
	}
}
