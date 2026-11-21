package iossim

import (
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func testAppParams() *app.Parameter {
	return &app.Parameter{Identifier: "com.example.game", App: &app.AppParams{AppPath: "/builds/Game.app"}}
}

func TestAppActionsRunSimctl(t *testing.T) {
	tests := []struct {
		name string
		call func(d *Device) error
		line string
	}{
		{"install", func(d *Device) error { return d.InstallApp(testAppParams()) }, "xcrun simctl install AAAA-1111 /builds/Game.app"},
		{"uninstall", func(d *Device) error { return d.UninstallApp(testAppParams()) }, "xcrun simctl uninstall AAAA-1111 com.example.game"},
		{"launch", func(d *Device) error { return d.StartApp(nil, testAppParams(), "s1", "http://node") }, "xcrun simctl launch AAAA-1111 com.example.game SESSION_ID s1 NODE_URL http://node"},
		{"terminate", func(d *Device) error { return d.StopApp(testAppParams()) }, "xcrun simctl terminate AAAA-1111 com.example.game"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testkit.FakeRunner{}
			r.Script(tt.line, nil, nil)
			if err := tt.call(newTestDevice(r, false)); err != nil {
				t.Fatal(err)
			}
			if got := r.Commands(); len(got) != 1 || got[0] != tt.line {
				t.Fatalf("commands: %v", got)
			}
			r.Script(tt.line, nil, errBoom)
			if err := tt.call(newTestDevice(r, false)); !errors.Is(err, errBoom) {
				t.Fatalf("error not propagated: %v", err)
			}
		})
	}
}

func TestKeepAppRunningSkipsLaunchAndTerminate(t *testing.T) {
	r := &testkit.FakeRunner{}
	d := newTestDevice(r, true)
	if err := d.StartApp(nil, testAppParams(), "s", "n"); err != nil {
		t.Fatal(err)
	}
	if err := d.StopApp(testAppParams()); err != nil {
		t.Fatal(err)
	}
	if got := r.Commands(); len(got) != 0 {
		t.Fatalf("no command expected, got %v", got)
	}
}

func TestIsAppInstalled(t *testing.T) {
	line := "xcrun simctl get_app_container AAAA-1111 com.example.game"
	r := &testkit.FakeRunner{}
	r.Script(line, nil, nil)
	if ok, err := newTestDevice(r, false).IsAppInstalled(testAppParams()); !ok || err != nil {
		t.Fatalf("installed: %v %v", ok, err)
	}
	r.Script(line, nil, errBoom)
	if ok, err := newTestDevice(r, false).IsAppInstalled(testAppParams()); ok || err == nil {
		t.Fatalf("not installed: %v %v", ok, err)
	}
}

// IsAppConnected was inverted once: it must be true only with a connection.
func TestIsAppConnected(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}, false)
	if d.IsAppConnected() {
		t.Fatal("no connection yet")
	}
	d.SetConnection(&device.Connection{ConnectionParameter: &appproto.Connect{}})
	if !d.IsAppConnected() {
		t.Fatal("connected app expected")
	}
	d.SetConnection(nil)
	if d.IsAppConnected() {
		t.Fatal("connection was dropped")
	}
}

func TestIdentityAndDefaults(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}, false)
	d.State = device.StateBooted
	if d.DeviceID() != udid1 || d.DeviceState() != device.StateBooted {
		t.Fatalf("identity: %q %v", d.DeviceID(), d.DeviceState())
	}
	if d.DeviceType() != int(models.DeviceTypePhone) || d.PlatformType() != int(models.PlatformTypeiOSSimulator) {
		t.Fatalf("types: %d %d", d.DeviceType(), d.PlatformType())
	}
	if d.ConnectionTimeout() != 10*time.Second || d.HasFeature("x") || !d.LastUpdate().IsZero() {
		t.Fatal("unexpected defaults")
	}
	d.State = parseSimState("nonsense")
	if d.DeviceState() != device.StateUnknown {
		t.Fatalf("unknown state expected, got %v", d.DeviceState())
	}
}

func TestUnsupportedActionsDoNotPanic(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}, false)
	d.Execute("feature")
	d.RunNativeScript([]byte("x"))
	if err := unsupported("x"); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("not ErrUnsupported: %v", err)
	}
}

func TestDefaultDeps(t *testing.T) {
	d := Deps{}.withDefaults()
	if _, ok := d.Runner.(device.ExecRunner); !ok {
		t.Fatalf("runner: %T", d.Runner)
	}
	if _, ok := d.Clock.(device.RealClock); !ok {
		t.Fatalf("clock: %T", d.Clock)
	}
}
