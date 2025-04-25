package events

var NewTestProtocolLog newTestProtocolLog

type NewTestProtocolLogPayload struct {
	TestProtocolID uint        `json:"testProtocolId"`
	Entry          interface{} `json:"entry"`
}

type newTestProtocolLog struct {
	handlers []interface {
		Handle(NewTestProtocolLogPayload)
	}
}

func (u *newTestProtocolLog) Register(handler interface {
	Handle(NewTestProtocolLogPayload)
}) {
	u.handlers = append(u.handlers, handler)
}

func (u newTestProtocolLog) Trigger(payload NewTestProtocolLogPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
