package events

var DeviceStatusChanged deviceStatusChanged

type DeviceStatusChangedPayload struct {
	DeviceID    uint `json:"deviceId"`
	DeviceState uint `json:"deviceState"`
}

type deviceStatusChanged struct {
	handlers []interface {
		Handle(DeviceStatusChangedPayload)
	}
}

func (u *deviceStatusChanged) Register(handler interface {
	Handle(DeviceStatusChangedPayload)
}) {
	u.handlers = append(u.handlers, handler)
}

func (u deviceStatusChanged) Trigger(payload DeviceStatusChangedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
