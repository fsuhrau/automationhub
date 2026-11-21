package mcp

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const sceneXML = `<Root ID="1" Name="root"><Button ID="8" Name="Play"/><Button ID="9" Name="Quit"/></Root>`

// appWorld is a world whose device has a connected app with a scene and a
// screenshot.
func appWorld(t *testing.T) (*goldenWorld, *mcp.ClientSession) {
	t.Helper()
	w := newGoldenWorld(t)
	w.dev.AppConnected = true
	w.pool.scene = sceneXML
	w.pool.screenshot = testPNG(t, 640, 480)
	cs := connect(t, w.serve(t, w.handler()), "alice")
	t.Cleanup(func() { _ = cs.Close() })
	return w, cs
}

func TestHierarchyOfTheConnectedApp(t *testing.T) {
	_, cs := appWorld(t)
	tests := []struct {
		name     string
		args     map[string]any
		contains string
		not      string
	}{
		{"whole scene", a("deviceId", 1), `Name="Quit"`, ""},
		{"by xpath", a("deviceId", 1, "xpath", "//Button[@Name='Play']"), "1 matching elements", `Name="Quit"`},
		{"truncated", a("deviceId", 1, "maxChars", 10), "... truncated", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := call(t, cs, "get_hierarchy", tc.args)

			got := text(res)
			if res.IsError || !strings.Contains(got, tc.contains) || (tc.not != "" && strings.Contains(got, tc.not)) {
				t.Fatalf("%v: %s", res.IsError, got)
			}
		})
	}
	if res := call(t, cs, "get_hierarchy", a("deviceId", 1, "xpath", "//[")); !res.IsError || !strings.Contains(text(res), "invalid xpath") {
		t.Errorf("invalid xpath: %s", text(res))
	}
}

func TestScreenshotOfTheConnectedApp(t *testing.T) {
	_, cs := appWorld(t)

	res := call(t, cs, "screenshot", a("deviceId", 1, "source", "app"))

	if res.IsError || !strings.Contains(text(res), "device screen 640x480") {
		t.Fatalf("%v %s", res.IsError, text(res))
	}
}

func TestTapOfAnElementOfTheConnectedApp(t *testing.T) {
	_, cs := appWorld(t)
	call(t, cs, "lock_device", a("deviceId", 1))

	byID := call(t, cs, "tap", a("deviceId", 1, "elementId", "8"))
	byXPath := call(t, cs, "tap", a("deviceId", 1, "xpath", "//Button[@Name='Play']"))

	if byID.IsError || text(byID) != "touched element 8" || byXPath.IsError || !strings.Contains(text(byXPath), "//Button") {
		t.Fatalf("by id %q, by xpath %q", text(byID), text(byXPath))
	}
}

func TestControlToolsReportAFailingDevice(t *testing.T) {
	tests := []struct {
		tool   string
		args   map[string]any
		method string
	}{
		{"screenshot", a("deviceId", 1), "GetScreenshot"},
		{"device_logs", a("deviceId", 1), "ReadLogs"},
		{"tap", a("deviceId", 1, "x", 1, "y", 2), "Tap"},
		{"swipe", a("deviceId", 1, "fromX", 1, "fromY", 1, "toX", 2, "toY", 2), "Swipe"},
		{"type_text", a("deviceId", 1, "text", "x"), "InputText"},
		{"press_key", a("deviceId", 1, "key", "home"), "PressKey"},
		{"install_app", a("deviceId", 1, "bundleId", 1), "InstallApp"},
		{"start_app", a("deviceId", 1, "bundleId", 1), "StartApp"},
		{"stop_app", a("deviceId", 1, "bundleId", 1), "StopApp"},
	}
	for _, tc := range tests {
		t.Run(tc.tool, func(t *testing.T) {
			srv, w := setup(t)
			cs := connect(t, srv, "alice")
			call(t, cs, "lock_device", a("deviceId", 1))
			w.dev.FailWith(tc.method, errBoom)

			res := call(t, cs, tc.tool, tc.args)

			if !res.IsError || !strings.Contains(text(res), "boom") {
				t.Fatalf("%v %q, want the device error", res.IsError, text(res))
			}
		})
	}
}

func TestStopAppClearsTheDataWhenAsked(t *testing.T) {
	srv, w := setup(t)
	cs := connect(t, srv, "alice")
	call(t, cs, "lock_device", a("deviceId", 1))

	call(t, cs, "stop_app", a("deviceId", 1, "bundleId", 1, "clearData", true))

	if len(w.dev.CallsTo("StopApp")) != 1 || len(w.dev.CallsTo("ClearAppData")) != 1 {
		t.Fatalf("calls %v", w.dev.Calls())
	}
}

func TestDeviceLogsAreCappedAtTheMaximum(t *testing.T) {
	srv, w := setup(t)
	cs := connect(t, srv, "alice")

	call(t, cs, "device_logs", a("deviceId", 1, "lines", 99999))

	if got := w.dev.CallsTo("ReadLogs"); len(got) != 1 || got[0].Args[0] != maxLogLines {
		t.Fatalf("ReadLogs calls %v, want %d lines", got, maxLogLines)
	}
}

// widePNG is a valid PNG too wide for JPEG.
func widePNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, 70000, 1))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestScreenshotsThatCannotBeEncodedAreErrors(t *testing.T) {
	srv, w := setup(t)
	w.dev.PNG, w.dev.Width, w.dev.Height = widePNG(t), 0, 0
	cs := connect(t, srv, "alice")
	protocol := shotsProtocol(t, w)
	if err := os.WriteFile(filepath.Join(w.dataDir, "shot-last.png"), widePNG(t), 0o644); err != nil {
		t.Fatal(err)
	}

	device := call(t, cs, "screenshot", a("deviceId", 1, "maxWidth", 100000))
	stored := call(t, cs, "get_protocol_screenshot", a("protocolId", protocol, "maxWidth", 100000))

	if !device.IsError || !stored.IsError {
		t.Fatalf("device screenshot error %v, protocol screenshot error %v", device.IsError, stored.IsError)
	}
}

func TestScreenshotOfADeviceThatReportsNoSizeUsesTheImage(t *testing.T) {
	srv, w := setup(t)
	w.dev.Width, w.dev.Height = 0, 0
	cs := connect(t, srv, "alice")

	res := call(t, cs, "screenshot", a("deviceId", 1))

	if !strings.Contains(text(res), "device screen 2000x1000") {
		t.Fatalf("%q", text(res))
	}
}

func TestScreenshotOfABrokenImageIsAnError(t *testing.T) {
	srv, w := setup(t)
	w.dev.PNG = []byte("not an image")
	cs := connect(t, srv, "alice")

	res := call(t, cs, "screenshot", a("deviceId", 1))

	if !res.IsError || !strings.Contains(text(res), "decode screenshot") {
		t.Fatalf("%q", text(res))
	}
}

func TestProtocolScreenshotReadErrors(t *testing.T) {
	srv, w := setup(t)
	if err := os.Mkdir(filepath.Join(w.dataDir, "dir.png"), 0o755); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, srv, "alice")
	protocol := shotsProtocol(t, w)
	w.db.Create(&models.ProtocolEntry{TestProtocolID: protocol, Data: "dir.png"})

	res := call(t, cs, "get_protocol_screenshot", a("protocolId", protocol, "file", "dir.png"))

	if !res.IsError {
		t.Fatalf("a folder was served as a screenshot: %q", text(res))
	}
}
