package events

import (
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

// lockDevice implements only the lock part of device.Device.
type lockDevice struct {
	device.Device
	id      string
	locked  bool
	lockErr error
}

func (d *lockDevice) DeviceID() string { return d.id }
func (d *lockDevice) IsLocked() bool   { return d.locked }
func (d *lockDevice) Lock() error {
	if d.lockErr != nil {
		return d.lockErr
	}
	d.locked = true
	return nil
}
func (d *lockDevice) Unlock() error {
	if !d.locked {
		return errors.New("not locked")
	}
	d.locked = false
	return nil
}

type lockRecorder chan DeviceLockChangedPayload

func (r lockRecorder) Handle(p DeviceLockChangedPayload) { r <- p }

func expectLockEvent(t *testing.T, r lockRecorder, want DeviceLockChangedPayload) {
	t.Helper()
	select {
	case got := <-r:
		if got != want {
			t.Fatalf("event = %+v, want %+v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("no lock event for %+v", want)
	}
}

func expectNoLockEvent(t *testing.T, r lockRecorder) {
	t.Helper()
	select {
	case got := <-r:
		t.Fatalf("unexpected event %+v", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestLockAndUnlockDeviceAnnounceTheChange(t *testing.T) {
	r := make(lockRecorder, 4)
	DeviceLockChanged.Register(r)
	dev := &lockDevice{id: "phone"}

	if err := LockDevice(dev); err != nil {
		t.Fatal(err)
	}
	expectLockEvent(t, r, DeviceLockChangedPayload{DeviceIdentifier: "phone", Locked: true})

	if err := UnlockDevice(dev); err != nil {
		t.Fatal(err)
	}
	expectLockEvent(t, r, DeviceLockChangedPayload{DeviceIdentifier: "phone", Locked: false})

	// nothing changed, so nothing is announced
	if err := UnlockDevice(dev); err == nil {
		t.Fatal("unlocking an unlocked device should fail")
	}
	dev.lockErr = errors.New("busy")
	if err := LockDevice(dev); err == nil {
		t.Fatal("lock error should be returned")
	}
	expectNoLockEvent(t, r)
}
