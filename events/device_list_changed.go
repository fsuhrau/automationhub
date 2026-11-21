package events

// DeviceListChanged is triggered when a device was added or its name, OS or
// parameters changed. Status changes are DeviceStatusChanged.
var DeviceListChanged = NewBus[DeviceListChangedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type DeviceListChangedPayload struct {
	DeviceID uint `json:"deviceId"`
	Added    bool `json:"added"`
}
