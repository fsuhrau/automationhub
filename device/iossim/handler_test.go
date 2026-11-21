package iossim

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func newHandler(store *memoryStore) (*Handler, *testkit.FakeRunner, *testkit.FakeClock) {
	r := &testkit.FakeRunner{}
	c := testkit.NewFakeClock(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	return NewHandlerWithDeps(store, Deps{Runner: r, Clock: c}), r, c
}

func TestNewHandlerUsesDefaults(t *testing.T) {
	h := NewHandler(&memoryStore{})
	if h.Name() != Manager || h.deps.Runner == nil {
		t.Fatal("handler not set up")
	}
}

func TestInitLoadsStoredSimulators(t *testing.T) {
	store := &memoryStore{devices: models.Devices{{DeviceIdentifier: udid1, Name: "iPhone 15", OS: "iphonesimulator", OSVersion: "17.0"}}}
	h, _, _ := newHandler(store)
	if err := h.Init("http://master", "node", nil); err != nil {
		t.Fatal(err)
	}
	d, ok := h.Get(udid1)
	if !ok || d.DeviceName() != "iPhone 15" || d.DeviceState() != device.StateUnknown || len(d.ActionHandlers()) != 1 {
		t.Fatalf("stored simulator not loaded: %+v", d)
	}
	store.getErr = errBoom
	if err := h.Init("m", "n", nil); !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
}

func TestStartStopRunCommands(t *testing.T) {
	h, r, _ := newHandler(&memoryStore{})
	open := "open -a " + simulatorApp
	r.Script(open, nil, nil)
	r.Script("killall Simulator", nil, errBoom)
	if err := h.Start(); err != nil {
		t.Fatal(err)
	}
	if err := h.Stop(); !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
	if got := strings.Join(r.Commands(), "|"); got != open+"|killall Simulator" {
		t.Fatalf("commands: %s", got)
	}
}

func TestBootAndShutdownKnownSimulator(t *testing.T) {
	h, r, _ := newHandler(&memoryStore{})
	h.Add(newTestDevice(r, false))
	r.Script("xcrun simctl boot AAAA-1111", nil, nil)
	r.Script("xcrun simctl shutdown AAAA-1111", nil, errBoom)
	if err := h.StartDevice(udid1); err != nil {
		t.Fatal(err)
	}
	if err := h.StopDevice(udid1); !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
	if got := strings.Join(r.Commands(), "|"); got != "xcrun simctl boot AAAA-1111|xcrun simctl shutdown AAAA-1111" {
		t.Fatalf("commands: %s", got)
	}
}

func TestBootUnknownSimulator(t *testing.T) {
	h, r, _ := newHandler(&memoryStore{})
	for _, err := range []error{h.StartDevice("nope"), h.StopDevice("nope")} {
		if !errors.Is(err, device.DeviceNotFoundError) {
			t.Fatalf("got %v", err)
		}
	}
	if len(r.Commands()) != 0 {
		t.Fatalf("no command expected: %v", r.Commands())
	}
}

func TestRegisterDeviceUnsupported(t *testing.T) {
	h, _, _ := newHandler(&memoryStore{})
	if _, err := h.RegisterDevice(device.RegisterData{}); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}
