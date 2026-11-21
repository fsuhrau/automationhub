package hub

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// failingRun is a device manager whose Run fails.
type failingRun struct{ *testkit.FakeDeviceManager }

func (failingRun) Run(context.Context, bool) error { return errors.New("no devices") }

// recordingPart is a background part that records the context it ran with.
type recordingPart struct{ ctx context.Context }

func (p *recordingPart) Run(ctx context.Context) { p.ctx = ctx }

func busyPort(t *testing.T) int32 {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return int32(l.Addr().(*net.TCPAddr).Port)
}

func serviceOn(t *testing.T, devices *testkit.FakeDeviceManager, port int32) *Service {
	t.Helper()
	s := newTestService(t, config.Service{Port: port}, Auth{})
	s.deviceManager = devices
	return s
}

func TestRunMasterStartsTheBackgroundPartsAndStopsThemWhenTheServerFails(t *testing.T) {
	devices := &testkit.FakeDeviceManager{}
	part := &recordingPart{}

	err := serviceOn(t, devices, busyPort(t)).RunMaster(context.Background(), part)

	if err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("err = %v", err)
	}
	if part.ctx == nil || part.ctx.Err() == nil {
		t.Errorf("the part must run and be stopped on return: %v", part.ctx)
	}
	if len(devices.CallsTo("Run")) != 1 || len(devices.CallsTo("StopObserver")) != 1 {
		t.Errorf("calls %v", devices.Calls())
	}
}

func TestRunMasterReturnsTheErrorOfTheDeviceManager(t *testing.T) {
	s := newTestService(t, config.Service{}, Auth{})
	s.deviceManager = failingRun{&testkit.FakeDeviceManager{}}

	if err := s.RunMaster(context.Background()); err == nil || err.Error() != "no devices" {
		t.Fatalf("err = %v", err)
	}
}

func TestRunMasterListensOnTheDefaultPortWithoutAConfiguredOne(t *testing.T) {
	s := serviceOn(t, &testkit.FakeDeviceManager{}, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	l, err := net.Listen("tcp", ":8002")
	if err != nil {
		t.Skipf("the default port is taken: %v", err)
	}
	defer func() { _ = l.Close() }()

	if err := s.RunMaster(ctx); err == nil || s.cfg.Port != defaultMasterPort {
		t.Fatalf("err = %v, port %d", err, s.cfg.Port)
	}
}

func TestRunNodeObservesTheConnectionAndStopsWhenTheServerFails(t *testing.T) {
	devices := &testkit.FakeDeviceManager{}
	devices.Add("android_device", testkit.NewFakeDevice("d"))

	err := serviceOn(t, devices, busyPort(t)).RunNode(context.Background(), &node.ReconnectHandler{})

	if err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("err = %v", err)
	}
	if len(devices.CallsTo("Run")) != 1 || len(devices.CallsTo("StopObserver")) != 1 {
		t.Errorf("calls %v", devices.Calls())
	}
}

func TestRunNodeReturnsTheErrorOfTheDeviceManager(t *testing.T) {
	s := newTestService(t, config.Service{}, Auth{})
	s.deviceManager = failingRun{&testkit.FakeDeviceManager{}}

	if err := s.RunNode(context.Background(), &node.ReconnectHandler{}); err == nil || err.Error() != "no devices" {
		t.Fatalf("err = %v", err)
	}
}
