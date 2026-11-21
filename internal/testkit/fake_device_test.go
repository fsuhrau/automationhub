package testkit

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
)

type handler struct{ id int }

func (*handler) OnActionResponse(interface{}, *appproto.Response) {}

func TestFakeDeviceReportsConfiguredValues(t *testing.T) {
	d := NewFakeDevice("id1")
	d.OSName, d.OSVersion, d.Target, d.Type, d.Platform = "android", "14", "34", 1, 2
	d.Parameter, d.State, d.Timeout, d.AppConnected = map[string]string{"k": "v"}, device.StateNone, time.Second, true
	if d.DeviceID() != "id1" || d.DeviceName() != "fake" || d.DeviceOSName() != "android" || d.DeviceOSVersion() != "14" ||
		d.TargetVersion() != "34" || d.DeviceType() != 1 || d.PlatformType() != 2 || d.DeviceState() != device.StateNone ||
		d.DeviceParameter()["k"] != "v" || d.ConnectionTimeout() != time.Second || !d.IsAppConnected() {
		t.Errorf("unexpected values: %+v", d)
	}
}

func TestFakeDeviceRecordsAndFailsCalls(t *testing.T) {
	p := &app.Parameter{}
	d := NewFakeDevice("1")
	d.Installed = true
	d.PNG, d.Width, d.Height = []byte("png"), 3, 4
	ops := map[string]func() error{
		"IsAppInstalled": func() error { _, err := d.IsAppInstalled(p); return err },
		"InstallApp":     func() error { return d.InstallApp(p) },
		"UninstallApp":   func() error { return d.UninstallApp(p) },
		"ClearAppData":   func() error { return d.ClearAppData(p) },
		"StopApp":        func() error { return d.StopApp(p) },
		"StartApp":       func() error { return d.StartApp(&device.DeviceConfig{}, p, "s", "u") },
		"Send":           func() error { return d.Send([]byte("x")) },
		"StartRecording": func() error { return d.StartRecording("f") },
		"StopRecording":  d.StopRecording,
		"GetScreenshot":  func() error { _, _, _, err := d.GetScreenshot(); return err },
		"Tap":            func() error { return d.Tap(1, 2) },
		"Swipe":          func() error { return d.Swipe(1, 2, 3, 4, time.Second) },
		"InputText":      func() error { return d.InputText("t") },
		"PressKey":       func() error { return d.PressKey("home") },
		"ReadLogs":       func() error { _, err := d.ReadLogs(1); return err },
		"ReadLogsSince":  func() error { _, err := d.ReadLogsSince("s", 1); return err },
		"StartVideo":     func() error { _, err := d.StartVideo(context.Background(), device.VideoOptions{}); return err },
		"Lock":           d.Lock,
		"Unlock":         d.Unlock,
	}
	for name, op := range ops {
		if err := op(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		d.FailWith(name, errBoom)
		if err := op(); err != errBoom {
			t.Errorf("%s scripted error: %v", name, err)
		}
		if got := len(d.CallsTo(name)); got != 2 {
			t.Errorf("%s calls = %d", name, got)
		}
	}
}

func TestFakeDeviceQueryResults(t *testing.T) {
	d := NewFakeDevice("1")
	d.Installed, d.PNG, d.Width, d.Height = true, []byte("png"), 3, 4
	d.Features, d.LogLines = []string{"scene"}, []string{"a", "b", "c"}
	if ok, _ := d.IsAppInstalled(nil); !ok {
		t.Error("installed")
	}
	if png, w, h, _ := d.GetScreenshot(); string(png) != "png" || w != 3 || h != 4 {
		t.Error("screenshot")
	}
	if !d.HasFeature("scene") || d.HasFeature("other") {
		t.Error("features")
	}
	cases := []struct {
		n    int
		want []string
	}{{2, []string{"b", "c"}}, {0, []string{"a", "b", "c"}}, {9, []string{"a", "b", "c"}}}
	for _, c := range cases {
		if got, _ := d.ReadLogs(c.n); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ReadLogs(%d) = %v", c.n, got)
		}
	}
	if got, _ := d.ReadLogsSince("t", 1); !reflect.DeepEqual(got, []string{"c"}) {
		t.Errorf("ReadLogsSince = %v", got)
	}
	s, err := d.StartVideo(context.Background(), device.VideoOptions{})
	if err != nil || s.Meta().Width != 3 {
		t.Errorf("video meta %+v err %v", s, err)
	}
}

func TestFakeDeviceLockState(t *testing.T) {
	d := NewFakeDevice("1")
	_ = d.Lock()
	if !d.IsLocked() {
		t.Error("not locked")
	}
	_ = d.Unlock()
	if d.IsLocked() {
		t.Error("still locked")
	}
	d.FailWith("Lock", errBoom)
	_ = d.Lock()
	if d.IsLocked() {
		t.Error("failed lock must not lock")
	}
	d.FailWith("Lock", nil)
	_ = d.Lock()
	d.FailWith("Unlock", errBoom)
	_ = d.Unlock()
	if !d.IsLocked() {
		t.Error("failed unlock must keep the lock")
	}
}

func TestFakeDeviceConnectionAndContext(t *testing.T) {
	d := NewFakeDevice("1")
	d.Cancel() // before NewContext: no panic
	c := &device.Connection{}
	d.SetConnection(c)
	if d.Connection() != c {
		t.Error("connection")
	}
	ctx := d.NewContext()
	d.Cancel()
	if ctx.Err() == nil {
		t.Error("context not cancelled")
	}
	if len(d.CallsTo("SetConnection")) != 1 || len(d.CallsTo("Cancel")) != 2 {
		t.Error("calls not recorded")
	}
}

func TestFakeDeviceLogWriterAndHandlers(t *testing.T) {
	d := NewFakeDevice("1")
	if d.GetLogWriter() != nil {
		t.Error("writer")
	}
	d.SetLogWriter(nil)
	h1, h2 := &handler{1}, &handler{2}
	d.AddActionHandler(h1)
	d.AddActionHandler(h2)
	d.RemoveActionHandler(&handler{3}) // unknown: no-op
	d.RemoveActionHandler(h1)
	if got := d.ActionHandlers(); len(got) != 1 || got[0] != h2 {
		t.Errorf("handlers = %v", got)
	}
}

func TestFakeDeviceRecordsFireAndForgetCalls(t *testing.T) {
	d := NewFakeDevice("1")
	d.Execute("e")
	d.RunNativeScript([]byte("s"))
	d.Data("src", "data")
	d.SetKey("k", "v")
	d.Passed(true)
	d.LogPerformance("cp", 1, 2, 3, 4, 5, "o")
	d.Log("src", "f %d", 1)
	d.Error("src", "f")
	d.Exception("src", "f")
	for _, m := range []string{"Execute", "RunNativeScript", "Data", "SetKey", "Passed", "LogPerformance", "Log", "Error", "Exception"} {
		if len(d.CallsTo(m)) != 1 {
			t.Errorf("%s not recorded", m)
		}
	}
}
