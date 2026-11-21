package node

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

//nolint:gochecknoglobals // go test flag
var update = flag.Bool("update", false, "rewrite the golden files in testdata")

func wireParameter() *app.Parameter {
	return &app.Parameter{
		Platform: "android", Name: "Demo", Identifier: "com.demo", Version: "1.2.3",
		App: &app.AppParams{
			AppBinaryID: 7, AppPath: "/data/app.apk", Additional: "extra", Hash: "h1", Size: 4,
			Android:    &app.AndroidParams{LaunchActivity: "com.demo.Main"},
			Executable: &app.ExecutableParams{Executable: "demo.exe"},
		},
		Web: &app.WebParams{StartURL: "http://demo.test"},
	}
}

func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("wire format of %s changed\nwant:\n%s\ngot:\n%s", name, want, got)
	}
}

// TestWireFormat round-trips every RPCClient method through a real RPCNode and
// compares the requests on the wire with the golden files, so old and new
// nodes and masters stay compatible.
func TestWireFormat(t *testing.T) {
	p := wireParameter()
	cfg := &device.DeviceConfig{Connection: 2, IP: "10.0.0.1", Port: 5555, DeviceParameter: map[string]string{"k": "v"}}
	cases := map[string]func(c *RPCClient) error{
		"Ping":           func(c *RPCClient) error { return c.Ping() },
		"GetDevices":     func(c *RPCClient) error { _, err := c.GetDevices(); return err },
		"StartDevice":    func(c *RPCClient) error { return c.StartDevice("dev-1") },
		"StopDevice":     func(c *RPCClient) error { return c.StopDevice("dev-1") },
		"IsAppInstalled": func(c *RPCClient) error { _, err := c.IsAppInstalled("dev-1", p); return err },
		"IsAppUploaded":  func(c *RPCClient) error { _, err := c.IsAppUploaded(p); return err },
		"UploadApp":      func(c *RPCClient) error { return c.UploadApp(context.Background(), p) },
		"InstallApp":     func(c *RPCClient) error { return c.InstallApp("dev-1", p) },
		"UninstallApp":   func(c *RPCClient) error { return c.UninstallApp("dev-1", p) },
		"StartApp":       func(c *RPCClient) error { return c.StartApp("dev-1", cfg, p, "session-1", "node.test") },
		"StopApp":        func(c *RPCClient) error { return c.StopApp("dev-1", p) },
		"ClearAppData":   func(c *RPCClient) error { return c.ClearAppData("dev-1", p) },
		"Control": func(c *RPCClient) error {
			_, err := c.Control(device.ControlRequest{DeviceID: "dev-1", Op: device.ControlTap, X: 3, Y: 4})
			return err
		},
		"StartVideoStream": func(c *RPCClient) error {
			return c.StartVideoStream(device.VideoStreamRequest{DeviceID: "dev-1", StreamID: "s-1", Options: device.VideoOptions{MaxSize: 720}})
		},
		"IsConnected":       func(c *RPCClient) error { c.IsConnected("dev-1"); return nil },
		"GetScreenshot":     func(c *RPCClient) error { _, _, _, err := c.GetScreenshot("dev-1"); return err },
		"HasFeature":        func(c *RPCClient) error { c.HasFeature("dev-1", "lua"); return nil },
		"Execute":           func(c *RPCClient) error { c.Execute("dev-1", "cmd"); return nil },
		"ConnectionTimeout": func(c *RPCClient) error { c.ConnectionTimeout("dev-1"); return nil },
		"RunNativeScript":   func(c *RPCClient) error { c.RunNativeScript("dev-1", []byte("script")); return nil },
		"SendAction":        func(c *RPCClient) error { c.SendAction("dev-1", []byte("action")); return nil },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			w := newWire(t)
			w.dev.PNG, w.dev.Width, w.dev.Height = []byte("png-bytes"), 9, 16
			if err := call(w.client); err != nil {
				t.Fatal(err)
			}
			checkGolden(t, name, w.rec.golden())
		})
	}
}

// The node has no recording RPC: the client doesn't pretend to record, so the
// node manager answers device.ErrUnsupported.
func TestRPCClientOffersTheRolesItsProtocolSupports(t *testing.T) {
	var client manager.RPCClient = &RPCClient{}
	checks := map[string]bool{
		"RPCScreenshotter":  assertsRole[manager.RPCScreenshotter](client),
		"RPCFeatures":       assertsRole[manager.RPCFeatures](client),
		"RPCNativeScripter": assertsRole[manager.RPCNativeScripter](client),
	}
	for role, ok := range checks {
		if !ok {
			t.Errorf("RPCClient should implement %s", role)
		}
	}
	if assertsRole[manager.RPCRecorder](client) {
		t.Error("RPCClient has no recording protocol, it must not implement RPCRecorder")
	}
}

func assertsRole[R any](client manager.RPCClient) bool {
	_, ok := client.(R)
	return ok
}
