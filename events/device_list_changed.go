package events

// DeviceListChanged is triggered when a device was added or its name, OS or
// parameters changed. Status changes are DeviceStatusChanged.
var DeviceListChanged deviceListChanged

type DeviceListChangedPayload struct {
	DeviceID uint `json:"deviceId"`
	Added    bool `json:"added"`
}

type deviceListChanged struct {
	handlers []interface {
		Handle(DeviceListChangedPayload)
	}
}

func (u *deviceListChanged) Register(handler interface {
	Handle(DeviceListChangedPayload)
}) {
	u.handlers = append(u.handlers, handler)
}

func (u deviceListChanged) Trigger(payload DeviceListChangedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
