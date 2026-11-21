package api

import (
	"github.com/fsuhrau/automationhub/events"
)

// RegisterEventDeviceLockListener asks the UI to reload the device list when a
// device was locked or unlocked, the same way a change of the live controller
// does. The list carries isLocked and lockedBy.
func RegisterEventDeviceLockListener(s *Service) {
	events.DeviceLockChanged.Register(events.HandlerFunc[events.DeviceLockChangedPayload](func(payload events.DeviceLockChangedPayload) {
		s.onDeviceControlChanged(payload.DeviceIdentifier)
	}))
}
