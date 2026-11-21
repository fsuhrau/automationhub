package events

var DeviceStatusChanged = NewBus[DeviceStatusChangedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type DeviceStatusChangedPayload struct {
	DeviceID    uint `json:"deviceId"`
	DeviceState uint `json:"deviceState"`
}
