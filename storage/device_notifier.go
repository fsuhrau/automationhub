package storage

import "github.com/fsuhrau/automationhub/events"

// DeviceNotifier hears about changes of the stored devices.
type DeviceNotifier interface {
	DeviceListChanged(events.DeviceListChangedPayload)
	DeviceStatusChanged(events.DeviceStatusChangedPayload)
}

// busNotifier forwards the changes to the event bus.
type busNotifier struct{}

func (busNotifier) DeviceListChanged(p events.DeviceListChangedPayload) {
	events.DeviceListChanged.Trigger(p)
}

func (busNotifier) DeviceStatusChanged(p events.DeviceStatusChangedPayload) {
	events.DeviceStatusChanged.Trigger(p)
}
