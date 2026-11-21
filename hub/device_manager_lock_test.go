package hub

import (
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

type lockRecorder chan events.DeviceLockChangedPayload

func (r lockRecorder) Handle(p events.DeviceLockChangedPayload) { r <- p }

func TestLockAndUnlockTriggerTheBusEvent(t *testing.T) {
	bus := events.NewBus[events.DeviceLockChangedPayload]()
	rec := make(lockRecorder, 4)
	bus.Register(rec)
	dm := newTestManager(t, WithLockEvents(bus))
	dev := testkit.NewFakeDevice("phone")

	steps := []struct {
		name string
		do   func() error
		want events.DeviceLockChangedPayload
	}{
		{"lock", func() error { return dm.LockDevice(dev) }, events.DeviceLockChangedPayload{DeviceIdentifier: "phone", Locked: true}},
		{"unlock", func() error { return dm.UnlockDevice(dev) }, events.DeviceLockChangedPayload{DeviceIdentifier: "phone", Locked: false}},
	}
	for _, s := range steps {
		if err := s.do(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		select {
		case got := <-rec:
			if got != s.want {
				t.Fatalf("%s: event %+v, want %+v", s.name, got, s.want)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s: no event", s.name)
		}
	}
}

func TestLockFailureAnnouncesNothing(t *testing.T) {
	bus := events.NewBus[events.DeviceLockChangedPayload]()
	rec := make(lockRecorder, 1)
	bus.Register(rec)
	dm := newTestManager(t, WithLockEvents(bus))
	dev := testkit.NewFakeDevice("phone")
	dev.FailWith("Lock", errors.New("busy"))
	dev.FailWith("Unlock", errors.New("not locked"))

	if dm.LockDevice(dev) == nil || dm.UnlockDevice(dev) == nil {
		t.Fatal("device errors must be returned")
	}
	select {
	case got := <-rec:
		t.Fatalf("unexpected event %+v", got)
	case <-time.After(50 * time.Millisecond):
	}
}
