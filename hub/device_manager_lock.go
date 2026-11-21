package hub

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
)

// LockDevice locks dev and announces the change on the lock bus, so the
// device pool shows the lock without a reload.
func (dm *DeviceManager) LockDevice(dev device.Device) error {
	return events.SetDeviceLock(dm.lockEvents, dev, true)
}

// UnlockDevice unlocks dev and announces the change.
func (dm *DeviceManager) UnlockDevice(dev device.Device) error {
	return events.SetDeviceLock(dm.lockEvents, dev, false)
}
