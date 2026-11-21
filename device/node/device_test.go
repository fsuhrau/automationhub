package node

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func newForwardingDevice(t *testing.T, nodes *testkit.FakeNodes) *NodeDevice {
	t.Helper()
	h, _ := newHandlerFor(t, nodes)
	d := h.newNodeDevice("node-1")
	d.ID = "phone"
	return d
}

func TestNodeDevice_ForwardsEveryCallToTheNode(t *testing.T) {
	p := &app.Parameter{Identifier: "com.demo"}
	cases := map[string]struct {
		call func(d *NodeDevice)
		want []any // arguments after the node and device id
	}{
		"IsAppInstalled":    {func(d *NodeDevice) { _, _ = d.IsAppInstalled(p) }, []any{p}},
		"InstallApp":        {func(d *NodeDevice) { _ = d.InstallApp(p) }, []any{p}},
		"UninstallApp":      {func(d *NodeDevice) { _ = d.UninstallApp(p) }, []any{p}},
		"StopApp":           {func(d *NodeDevice) { _ = d.StopApp(p) }, []any{p}},
		"ClearAppData":      {func(d *NodeDevice) { _ = d.ClearAppData(p) }, []any{p}},
		"IsConnected":       {func(d *NodeDevice) { d.IsAppConnected() }, nil},
		"StartRecording":    {func(d *NodeDevice) { _ = d.StartRecording("/tmp/x") }, []any{"/tmp/x"}},
		"StopRecording":     {func(d *NodeDevice) { _ = d.StopRecording() }, nil},
		"GetScreenshot":     {func(d *NodeDevice) { _, _, _, _ = d.GetScreenshot() }, nil},
		"HasFeature":        {func(d *NodeDevice) { d.HasFeature("lua") }, []any{"lua"}},
		"Execute":           {func(d *NodeDevice) { d.Execute("cmd") }, []any{"cmd"}},
		"ConnectionTimeout": {func(d *NodeDevice) { d.ConnectionTimeout() }, nil},
		"RunNativeScript":   {func(d *NodeDevice) { d.RunNativeScript([]byte("s")) }, []any{[]byte("s")}},
		"SendAction":        {func(d *NodeDevice) { _ = d.Send([]byte("a")) }, []any{[]byte("a")}},
	}
	for method, tc := range cases {
		t.Run(method, func(t *testing.T) {
			nodes := &testkit.FakeNodes{}
			tc.call(newForwardingDevice(t, nodes))
			calls := nodes.CallsTo(method)
			if len(calls) != 1 {
				t.Fatalf("calls = %v", nodes.Calls())
			}
			want := append([]any{manager.NodeIdentifier("node-1"), "phone"}, tc.want...)
			if !reflect.DeepEqual(calls[0].Args, want) {
				t.Fatalf("args = %v, want %v", calls[0].Args, want)
			}
		})
	}
}

func TestNodeDevice_ReturnsNodeAnswersAndErrors(t *testing.T) {
	boom := errors.New("node said no")
	nodes := &testkit.FakeNodes{Installed: true, Connected: true, Feature: true, Timeout: time.Minute, PNG: []byte("png"), Width: 2, Height: 3}
	d := newForwardingDevice(t, nodes)
	if ok, err := d.IsAppInstalled(&app.Parameter{}); !ok || err != nil {
		t.Errorf("IsAppInstalled = %v, %v", ok, err)
	}
	if !d.IsAppConnected() || !d.HasFeature("f") || d.ConnectionTimeout() != time.Minute {
		t.Error("queries")
	}
	if png, w, h, err := d.GetScreenshot(); string(png) != "png" || w != 2 || h != 3 || err != nil {
		t.Errorf("screenshot = %q %d %d %v", png, w, h, err)
	}
	nodes.FailWith("InstallApp", boom)
	if err := d.InstallApp(&app.Parameter{}); !errors.Is(err, boom) {
		t.Errorf("InstallApp = %v", err)
	}
	if err := d.Send(nil); err != nil {
		t.Errorf("Send = %v", err)
	}
}

func TestNodeDevice_InputAndLogsGoThroughControl(t *testing.T) {
	nodes := &testkit.FakeNodes{ControlReply: device.ControlResponse{Lines: []string{"l1"}}}
	d := newForwardingDevice(t, nodes)
	_ = d.Tap(1, 2)
	_ = d.Swipe(1, 2, 3, 4, 250*time.Millisecond)
	_ = d.InputText("hi")
	_ = d.PressKey("back")
	lines, err := d.ReadLogs(5)
	if err != nil || len(lines) != 1 {
		t.Fatalf("ReadLogs = %v, %v", lines, err)
	}
	if _, err := d.ReadLogsSince("01-02 03:04:05.000", 7); err != nil {
		t.Fatal(err)
	}
	var got []device.ControlRequest
	for _, c := range nodes.CallsTo("Control") {
		got = append(got, c.Args[1].(device.ControlRequest))
	}
	want := []device.ControlRequest{
		{DeviceID: "phone", Op: device.ControlTap, X: 1, Y: 2},
		{DeviceID: "phone", Op: device.ControlSwipe, X: 1, Y: 2, ToX: 3, ToY: 4, DurationMs: 250},
		{DeviceID: "phone", Op: device.ControlText, Text: "hi"},
		{DeviceID: "phone", Op: device.ControlKey, Key: "back"},
		{DeviceID: "phone", Op: device.ControlReadLogs, Lines: 5},
		{DeviceID: "phone", Op: device.ControlReadLogs, Lines: 7, Since: "01-02 03:04:05.000"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("control requests\n got %+v\nwant %+v", got, want)
	}
}

func TestNodeDevice_StartAppSendsTheStoredConfig(t *testing.T) {
	nodes := &testkit.FakeNodes{}
	d := newForwardingDevice(t, nodes)
	d.SetConfig(&models.Device{
		ConnectionParameter: &models.ConnectionParameter{ConnectionType: 2, IP: "10.0.0.5", Port: 5555},
		CustomParameter:     []models.CustomParameter{{Key: "pin", Value: "1234"}},
	})
	ignored := &device.DeviceConfig{IP: "ignored"}
	if err := d.StartApp(ignored, &app.Parameter{}, "session", "node-url"); err != nil {
		t.Fatal(err)
	}
	args := nodes.CallsTo("StartApp")[0].Args
	want := &device.DeviceConfig{Connection: 2, IP: "10.0.0.5", Port: 5555, DeviceParameter: map[string]string{"pin": "1234"}}
	if !reflect.DeepEqual(args[2], want) || args[4] != "session" || args[5] != "node-url" {
		t.Fatalf("args = %v", args)
	}
}

func TestNodeDevice_StartAppWithoutStoredConfig(t *testing.T) {
	for name, cfg := range map[string]*models.Device{"no config": nil, "empty config": {}} {
		t.Run(name, func(t *testing.T) {
			nodes := &testkit.FakeNodes{}
			d := newForwardingDevice(t, nodes)
			d.SetConfig(cfg)
			_ = d.StartApp(nil, &app.Parameter{}, "s", "n")
			got := nodes.CallsTo("StartApp")[0].Args[2].(*device.DeviceConfig)
			if got.IP != "" || got.Connection != 0 || len(got.DeviceParameter) != 0 || got.DeviceParameter == nil {
				t.Fatalf("config = %+v", got)
			}
		})
	}
}

func TestNodeDevice_Identity(t *testing.T) {
	nodes := &testkit.FakeNodes{}
	d := newForwardingDevice(t, nodes)
	if d.GetNodeID() != "node-1" || d.NodeManager() != nodes || d.DeviceID() != "phone" {
		t.Error("identity")
	}
	for name, want := range map[string]device.State{
		"StateBooted": device.StateBooted, "StateShutdown": device.StateShutdown, "nonsense": device.StateUnknown,
	} {
		d.SetDeviceState(name)
		if d.DeviceState() != want {
			t.Errorf("%s: state = %v", name, d.DeviceState())
		}
	}
}

func TestNodeDevice_UpdateDeviceInfosCopies(t *testing.T) {
	d := newForwardingDevice(t, &testkit.FakeNodes{})
	in := map[string]string{"a": "1"}
	d.UpdateDeviceInfos(in)
	in["a"] = "changed"
	if got := d.DeviceParameter(); got["a"] != "1" || len(got) != 1 {
		t.Fatalf("parameter = %v", got)
	}
	d.UpdateDeviceInfos(nil)
	if got := d.DeviceParameter(); got == nil || len(got) != 0 {
		t.Fatalf("parameter = %#v", got)
	}
}

func TestNodeDevice_RefreshTakesOverTheReportedDevice(t *testing.T) {
	reported := testkit.NewFakeDevice("phone")
	reported.Name, reported.OSName, reported.OSVersion = "Pixel", "Android", "14"
	reported.Type, reported.Platform, reported.State = 1, 2, device.StateBooted
	reported.Parameter = map[string]string{"k": "v"}
	now := time.Now()
	d := newForwardingDevice(t, &testkit.FakeNodes{})
	d.refresh(reported, now)
	if d.DeviceName() != "Pixel" || d.DeviceOSName() != "Android" || d.DeviceOSVersion() != "14" ||
		d.DeviceType() != 1 || d.PlatformType() != 2 || d.DeviceState() != device.StateBooted ||
		d.DeviceParameter()["k"] != "v" || !d.lastUpdateAt.Equal(now) {
		t.Fatalf("device = %+v", d)
	}
}

// providerNodes owns a rendezvous like the real node manager.
type providerNodes struct {
	*testkit.FakeNodes
	rendezvous *videostream.Rendezvous
}

func (p providerNodes) Rendezvous() *videostream.Rendezvous { return p.rendezvous }

func TestNewHandler_UsesTheRendezvousOfTheNodeManager(t *testing.T) {
	shared := videostream.NewRendezvous()
	h, _ := newHandlerFor(t, providerNodes{&testkit.FakeNodes{}, shared})
	if h.rendezvous != shared || h.newNodeDevice("n").rendezvous != shared {
		t.Error("handler and devices must share the node manager's rendezvous")
	}
	own, _ := newHandlerFor(t, &testkit.FakeNodes{})
	if own.rendezvous == nil || own.rendezvous == shared {
		t.Error("without a provider the handler needs a private rendezvous")
	}
}

func TestStartVideo_FailureAndTimeout(t *testing.T) {
	boom := errors.New("node refused")
	nodes := &testkit.FakeNodes{}
	nodes.FailWith("StartVideoStream", boom)
	d := newForwardingDevice(t, nodes)
	if _, err := d.StartVideo(context.Background(), device.VideoOptions{}); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}

	// the node accepts the request but never dials back
	nodes.FailWith("StartVideoStream", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := d.StartVideo(ctx, device.VideoOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if len(nodes.CallsTo("StartVideoStream")) != 2 {
		t.Error("the request must reach the node")
	}
}
