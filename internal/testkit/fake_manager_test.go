package testkit

import (
	"context"
	"sort"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

func TestFakeDeviceManagerServesAddedDevices(t *testing.T) {
	m := &FakeDeviceManager{}
	a, b := NewFakeDevice("a"), NewFakeDevice("b")
	m.Add("android", a)
	m.Add("ios", b)
	if d, name := m.GetDevice("b"); d != device.Device(b) || name != "ios" {
		t.Errorf("GetDevice b = %v %q", d, name)
	}
	if d, name := m.GetDevice("zz"); d != nil || name != "" {
		t.Errorf("unknown device = %v %q", d, name)
	}
	devs, err := m.Devices()
	if err != nil || len(devs["android"]) != 1 {
		t.Errorf("Devices = %v %v", devs, err)
	}
	names := m.Managers()
	sort.Strings(names)
	if len(names) != 2 || names[0] != "android" {
		t.Errorf("Managers = %v", names)
	}
}

func TestFakeDeviceManagerAnswersAndRecords(t *testing.T) {
	a := NewFakeDevice("a")
	m := &FakeDeviceManager{Found: a, Registered: a}
	m.FailWith("RegisterDevice", errBoom)
	if got := m.FindDevice(&app.Parameter{}); got != device.Device(a) {
		t.Error("FindDevice")
	}
	if got, err := m.RegisterDevice(device.RegisterData{}); got != device.Device(a) || err != errBoom {
		t.Errorf("RegisterDevice = %v %v", got, err)
	}
	calls := map[string]error{
		"Run":        m.Run(context.Background(), true),
		"Start":      m.Start(a),
		"Stop":       m.Stop(a),
		"SendAction": m.SendAction(a, nil),
	}
	m.AddHandler(nil)
	m.StopObserver()
	for name, err := range calls {
		if err != nil || len(m.CallsTo(name)) != 1 {
			t.Errorf("%s err=%v calls=%d", name, err, len(m.CallsTo(name)))
		}
	}
	if len(m.CallsTo("AddHandler")) != 1 || len(m.CallsTo("StopObserver")) != 1 {
		t.Error("AddHandler/StopObserver not recorded")
	}
}

func TestFakeNodesQueries(t *testing.T) {
	n := &FakeNodes{Status: 2, Uploaded: true, Installed: true, Connected: true, Feature: true,
		PNG: []byte("p"), Width: 1, Height: 2, Timeout: 5, ControlReply: device.ControlResponse{Lines: []string{"l"}},
		Managed: map[string][]device.Device{"m": nil}}
	n.RegisterNode("n1", nil, "h", "a", "os", "env", 1, nil)
	if !n.IsNodeKnown("n1") || n.IsNodeKnown("n2") || len(n.GetNodes()) != 1 {
		t.Error("node registry")
	}
	id := manager.NodeIdentifier("n1")
	p := &app.Parameter{}
	if s, _ := n.GetStatus(id); s != 2 {
		t.Error("status")
	}
	if got, _ := n.GetManagers(id); len(got) != 1 {
		t.Error("managers")
	}
	if ok, _ := n.IsAppUploaded(id, p); !ok {
		t.Error("uploaded")
	}
	if ok, _ := n.IsAppInstalled(id, "d", p); !ok {
		t.Error("installed")
	}
	if !n.IsConnected(id, "d") || !n.HasFeature(id, "d", "f") || n.ConnectionTimeout(id, "d") != 5 {
		t.Error("connected/feature/timeout")
	}
	if png, w, h, _ := n.GetScreenshot(id, "d"); string(png) != "p" || w != 1 || h != 2 {
		t.Error("screenshot")
	}
	if r, _ := n.Control(id, device.ControlRequest{}); len(r.Lines) != 1 {
		t.Error("control")
	}
}

func TestFakeNodesRecordsAndFailsActions(t *testing.T) {
	n := &FakeNodes{}
	id, p := manager.NodeIdentifier("n"), &app.Parameter{}
	ops := map[string]func() error{
		"StartDevice":      func() error { return n.StartDevice(id, "d") },
		"StopDevice":       func() error { return n.StopDevice(id, "d") },
		"UploadApp":        func() error { return n.UploadApp(context.Background(), id, p) },
		"InstallApp":       func() error { return n.InstallApp(id, "d", p) },
		"UninstallApp":     func() error { return n.UninstallApp(id, "d", p) },
		"StartApp":         func() error { return n.StartApp(id, "d", nil, p, "s", "u") },
		"StopApp":          func() error { return n.StopApp(id, "d", p) },
		"ClearAppData":     func() error { return n.ClearAppData(id, "d", p) },
		"StartVideoStream": func() error { return n.StartVideoStream(id, device.VideoStreamRequest{}) },
		"StartRecording":   func() error { return n.StartRecording(id, "d", "path") },
		"StopRecording":    func() error { return n.StopRecording(id, "d") },
	}
	for name, op := range ops {
		n.FailWith(name, errBoom)
		if err := op(); err != errBoom {
			t.Errorf("%s: %v", name, err)
		}
		if len(n.CallsTo(name)) != 1 {
			t.Errorf("%s not recorded", name)
		}
	}
	n.Run(context.Background())
	n.Execute(id, "d", "e")
	n.RunNativeScript(id, "d", nil)
	n.SendAction(id, "d", nil)
	for _, m := range []string{"Run", "Execute", "RunNativeScript", "SendAction"} {
		if len(n.CallsTo(m)) != 1 {
			t.Errorf("%s not recorded", m)
		}
	}
}
