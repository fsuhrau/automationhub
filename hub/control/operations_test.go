package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// leased is a service with a fake device and the target of its lease.
func leased(t *testing.T) (*Service, *testkit.FakeDevice, Target) {
	t.Helper()
	s := New(nil)
	dev := testkit.NewFakeDevice("d1")
	lease, err := s.Acquire(dev, "alice")
	if err != nil {
		t.Fatal(err)
	}
	return s, dev, Target{Device: dev, Token: lease.Token}
}

func TestInputCallsReachTheDevice(t *testing.T) {
	s, dev, target := leased(t)
	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"swipe", func() error { return s.Swipe(target, SwipeGesture{1, 2, 3, 4, time.Second}) }, "Swipe"},
		{"text", func() error { return s.InputText(target, "hello") }, "InputText"},
		{"key", func() error { return s.PressKey(target, "home") }, "PressKey"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Fatal(err)
			}
			if len(dev.CallsTo(tc.want)) != 1 {
				t.Fatalf("calls %v", dev.Calls())
			}
		})
	}
	if got := dev.CallsTo("Swipe")[0].Args; got[0] != 1 || got[3] != 4 || got[4] != time.Second {
		t.Fatalf("swipe args %v", got)
	}
}

func TestInputCallsNeedTheLeaseAndAnInputCapableDevice(t *testing.T) {
	s, _, target := leased(t)
	foreign := Target{Device: target.Device, Token: "wrong"}
	calls := map[string]func(Target) error{
		"swipe": func(t Target) error { return s.Swipe(t, SwipeGesture{}) },
		"text":  func(t Target) error { return s.InputText(t, "x") },
		"key":   func(t Target) error { return s.PressKey(t, "x") },
	}
	for name, call := range calls {
		if err := call(foreign); !errors.Is(err, ErrLeaseInvalid) {
			t.Errorf("%s: err %v, want ErrLeaseInvalid", name, err)
		}
	}
	plain := &bareDevice{id: "p1"}
	lease, _ := s.Acquire(plain, "alice")
	if err := s.Tap(Target{plain, lease.Token}, 1, 1); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("a device without input: %v", err)
	}
}

// bareDevice has the core roles only: no input, no data clearing.
type bareDevice struct {
	device.Device
	id string
}

func (d *bareDevice) DeviceID() string { return d.id }
func (*bareDevice) Lock() error        { return nil }
func (*bareDevice) Unlock() error      { return nil }

func TestInjectInputAuthorizesEverythingButAVideoReset(t *testing.T) {
	s, _, target := leased(t)
	stream := &inputStream{}
	reset := device.InputEvent{Kind: device.InputResetVideo}
	touch := device.InputEvent{Kind: device.InputTouch}

	if err := s.InjectInput(Target{Device: target.Device}, stream, reset); err != nil {
		t.Fatalf("reset without a lease: %v", err)
	}
	if err := s.InjectInput(Target{Device: target.Device}, stream, touch); !errors.Is(err, ErrLeaseRequired) {
		t.Fatalf("touch without a lease: %v", err)
	}
	if err := s.InjectInput(target, stream, touch); err != nil {
		t.Fatal(err)
	}
	if err := s.InjectInput(target, noInput{}, touch); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("a stream without input: %v", err)
	}
	if len(stream.events) != 2 {
		t.Fatalf("events %v", stream.events)
	}
}

type inputStream struct {
	noInput
	events []device.InputEvent
}

func (s *inputStream) InjectInput(ev device.InputEvent) error {
	s.events = append(s.events, ev)
	return nil
}

type noInput struct{}

func (noInput) Meta() device.VideoMeta                  { return device.VideoMeta{} }
func (noInput) ReadPacket() (device.VideoPacket, error) { return device.VideoPacket{}, nil }
func (noInput) Close() error                            { return nil }

func TestAppLifecycleCalls(t *testing.T) {
	s, dev, target := leased(t)
	params := &app.Parameter{Identifier: "com.x"}

	if err := s.InstallApp(context.Background(), target, params); err != nil {
		t.Fatal(err)
	}
	if err := s.StartApp(target, params, Launch{SessionID: "sess", NodeURL: "http://node"}); err != nil {
		t.Fatal(err)
	}
	if err := s.StopApp(target, params, false); err != nil {
		t.Fatal(err)
	}
	if err := s.StopApp(target, params, true); err != nil {
		t.Fatal(err)
	}

	start := dev.CallsTo("StartApp")[0].Args
	if start[2] != "sess" || start[3] != "http://node" {
		t.Fatalf("start args %v", start)
	}
	if len(dev.CallsTo("InstallApp")) != 1 || len(dev.CallsTo("StopApp")) != 2 || len(dev.CallsTo("ClearAppData")) != 1 {
		t.Fatalf("calls %v", dev.Calls())
	}
}

func TestAppLifecycleFailures(t *testing.T) {
	s, dev, target := leased(t)
	params := &app.Parameter{}
	boom := errors.New("boom")
	foreign := Target{Device: dev, Token: "wrong"}

	if err := s.InstallApp(context.Background(), foreign, params); !errors.Is(err, ErrLeaseInvalid) {
		t.Errorf("install without lease: %v", err)
	}
	if err := s.StartApp(foreign, params, Launch{}); !errors.Is(err, ErrLeaseInvalid) {
		t.Errorf("start without lease: %v", err)
	}
	if err := s.StopApp(foreign, params, false); !errors.Is(err, ErrLeaseInvalid) {
		t.Errorf("stop without lease: %v", err)
	}
	dev.FailWith("StopApp", boom)
	if err := s.StopApp(target, params, true); !errors.Is(err, boom) {
		t.Errorf("stop error: %v", err)
	}
	if len(dev.CallsTo("ClearAppData")) != 0 {
		t.Error("data cleared although the app did not stop")
	}
}

func TestStopAppWithoutDataClearingSupport(t *testing.T) {
	s := New(nil)
	plain := &stoppable{bareDevice: bareDevice{id: "p1"}}
	lease, _ := s.Acquire(plain, "alice")

	err := s.StopApp(Target{plain, lease.Token}, &app.Parameter{}, true)

	if !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("err %v", err)
	}
}

// stoppable can stop apps but doesn't clear their data.
type stoppable struct{ bareDevice }

func (*stoppable) StopApp(*app.Parameter) error { return nil }
