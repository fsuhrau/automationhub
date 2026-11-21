package unityeditor

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gorilla/websocket"
)

// editorServer is the hub side of the editor's websocket: it accepts the
// connection and reports the messages it receives.
func editorServer(t *testing.T) (*websocket.Conn, <-chan string) {
	t.Helper()
	received := make(chan string, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			received <- string(msg)
		}
	}))
	t.Cleanup(server.Close)
	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return client, received
}

// TestEditorLifecycleThroughTheHandler registers an editor over a websocket,
// finds it in the device list, starts and stops it, and sees it shut down once
// the websocket closes.
func TestEditorLifecycleThroughTheHandler(t *testing.T) {
	dir := project(t)
	conn, _ := editorServer(t)
	clock := newTickClock(t, func(time.Time) {
		_ = os.WriteFile(filepath.Join(dir, instanceFileName), nil, 0o644)
	})
	env := newEnv(clock)
	data := registerData("e1")
	data.ProjectDir, data.Conn = dir, conn

	registered, err := env.handler.RegisterDevice(data)
	if err != nil {
		t.Fatal(err)
	}
	devices, _ := env.handler.GetDevices()
	if len(devices) != 1 || devices[0] != registered {
		t.Fatalf("devices = %v, want the registered editor", devices)
	}
	eventually(t, "editor booted", func() bool { return registered.DeviceState() == device.StateBooted })

	env.runner.Script(launchLine(dir, epochTime()), nil, nil)
	if err := env.handler.StartDevice("e1"); err != nil {
		t.Fatalf("StartDevice: %v", err)
	}
	editor, _ := env.handler.Get("e1")
	process := editor.process.(*testkit.FakeProcess)

	if err := env.handler.StopDevice("e1"); err != nil {
		t.Fatalf("StopDevice: %v", err)
	}
	if process.Killed() != 1 {
		t.Fatalf("killed = %d, want 1", process.Killed())
	}
	_ = conn.Close()
	eventually(t, "editor shut down", func() bool { return registered.DeviceState() == device.StateShutdown })
}

// TestEditorReceivesStartAndStop shows the app commands reach the editor.
func TestEditorReceivesStartAndStop(t *testing.T) {
	conn, received := editorServer(t)
	env := newEnv(nil)
	data := registerData("e1")
	data.Conn = conn
	registered, _ := env.handler.RegisterDevice(data)
	eventually(t, "editor booted", func() bool { return registered.DeviceState() == device.StateBooted })

	if err := registered.StartApp(nil, nil, "s1", "node:1"); err != nil {
		t.Fatal(err)
	}
	if err := registered.StopApp(nil); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != `{"Action":"start","HostUrl":"node:1","SessionID":"s1"}` {
		t.Errorf("start = %s", got)
	}
	if got := <-received; got != `{"Action":"stop"}` {
		t.Errorf("stop = %s", got)
	}
}
