package events

import "github.com/fsuhrau/automationhub/device"

// DeviceLockChanged is triggered when a device was locked or unlocked, e.g.
// by a test run or the unlock button. The lock state is not stored, so the UI
// learns about it only through this event.
var DeviceLockChanged = NewBus[DeviceLockChangedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type DeviceLockChangedPayload struct {
	DeviceIdentifier string `json:"deviceIdentifier"`
	Locked           bool   `json:"locked"`
}

// NOTE(G30): moves to hub/device manager

// LockDevice locks dev and announces the change. Use it instead of dev.Lock
// so the device pool shows the lock without a reload.
func LockDevice(dev device.Device) error {
	if err := dev.Lock(); err != nil {
		return err
	}
	DeviceLockChanged.Trigger(DeviceLockChangedPayload{DeviceIdentifier: dev.DeviceID(), Locked: true})
	return nil
}

// UnlockDevice unlocks dev and announces the change.
func UnlockDevice(dev device.Device) error {
	if err := dev.Unlock(); err != nil {
		return err
	}
	DeviceLockChanged.Trigger(DeviceLockChangedPayload{DeviceIdentifier: dev.DeviceID(), Locked: false})
	return nil
}
