package api

import (
	"github.com/fsuhrau/automationhub/events"
)

// deviceLockNotifier asks the UI to reload the device list when a device was
// locked or unlocked, the same way a change of the live controller does. The
// list carries isLocked and lockedBy.
type deviceLockNotifier struct {
	s *Service
}

func RegisterEventDeviceLockListener(s *Service) {
	events.DeviceLockChanged.Register(deviceLockNotifier{s: s})
}

func (n deviceLockNotifier) Handle(payload events.DeviceLockChangedPayload) {
	n.s.onDeviceControlChanged(payload.DeviceIdentifier)
}
