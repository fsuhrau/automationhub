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

// SetDeviceLock locks or unlocks dev and announces the change on bus. Nothing
// is announced when the device refuses the change.
func SetDeviceLock(bus *Bus[DeviceLockChangedPayload], dev device.Device, locked bool) error {
	change := dev.Unlock
	if locked {
		change = dev.Lock
	}
	if err := change(); err != nil {
		return err
	}
	bus.Trigger(DeviceLockChangedPayload{DeviceIdentifier: dev.DeviceID(), Locked: locked})
	return nil
}

// LockDevice locks dev and announces the change.
//
// Deprecated: use hub.DeviceManager.LockDevice, which takes the bus it
// announces on from the manager.
func LockDevice(dev device.Device) error {
	return SetDeviceLock(DeviceLockChanged, dev, true)
}

// UnlockDevice unlocks dev and announces the change.
//
// Deprecated: use hub.DeviceManager.UnlockDevice.
func UnlockDevice(dev device.Device) error {
	return SetDeviceLock(DeviceLockChanged, dev, false)
}
