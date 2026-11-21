package mcp

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func testPNG(t *testing.T, w, h int) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setup(t *testing.T) (*httptest.Server, *goldenWorld) {
	t.Helper()
	w := newGoldenWorld(t)
	return w.serve(t, w.handler()), w
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
	t.Cleanup(func() { _ = cs.Close() })
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
	srv, w := setup(t)
	dev := w.dev
	alice := connect(t, srv, "alice")
	bob := connect(t, srv, "bob")

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
	if taps := dev.CallsTo("Tap"); len(taps) != 1 || taps[0].Args[0] != 5 || taps[0].Args[1] != 6 {
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
	res := call(t, cs, "get_device", map[string]any{"deviceId": 42})
	if !res.IsError || !strings.Contains(text(res), "not found") {
		t.Fatalf("expected not found, got %q", text(res))
	}
	res = call(t, cs, "list_devices", map[string]any{"platform": "android"})
	if res.IsError || !strings.Contains(text(res), "Pixel") {
		t.Fatalf("list_devices failed: %s", text(res))
	}
}

func TestTheHolderOfALockIsShownOnTheDevice(t *testing.T) {
	srv, _ := setup(t)
	cs := connect(t, srv, "alice")
	call(t, cs, "lock_device", a("deviceId", 1))

	got := text(call(t, cs, "get_device", a("deviceId", 1)))

	if !strings.Contains(got, `"locked": true`) || !strings.Contains(got, `"lockedBy": "mcp:alice:`) {
		t.Fatalf("device %s", got)
	}
}

// failingRelease is the live control whose lease cannot be released.
type failingRelease struct{ Control }

func (failingRelease) Release(string, string) error { return errBoom }

func TestReleaseReportsALeaseThatCannotBeReleased(t *testing.T) {
	w := newBareWorld(t)
	h := NewHandler(Deps{Pool: w.pool, Control: failingRelease{w.control}, Lookup: &stub{}, Tracker: w.tracker})
	cs := connect(t, w.serve(t, h), "alice")
	call(t, cs, "lock_device", a("deviceId", 1))

	res := call(t, cs, "release_device", a("deviceId", 1))

	if !res.IsError || text(res) != "boom" {
		t.Fatalf("%v %q", res.IsError, text(res))
	}
}

func TestJSONResultReportsAValueThatCannotBeEncoded(t *testing.T) {
	if _, _, err := jsonResult(math.NaN()); err == nil {
		t.Fatal("NaN was encoded")
	}
}

// cancelAfterReport is the run service whose first report cancels the call.
type cancelAfterReport struct {
	RunService
	cancel context.CancelFunc
}

func (c cancelAfterReport) Report(ctx context.Context, projectID, runID uint) (*runsvc.Report, error) {
	report, err := c.RunService.Report(ctx, projectID, runID)
	c.cancel()
	return report, err
}

func TestWaitEndsWithTheErrorOfACancelledCall(t *testing.T) {
	w := newGoldenWorld(t)
	w.tracker.Started(1)
	ctx, cancel := context.WithCancel(context.Background())
	deps := w.handler().deps
	deps.Runs = cancelAfterReport{deps.Runs, cancel}
	tl := &tools{Deps: deps, project: w.project}

	_, _, err := tl.waitForTestRun(ctx, &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}}, WaitArgs{RunID: 1})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error %v, want the cancellation", err)
	}
}
