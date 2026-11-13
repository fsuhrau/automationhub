package events

var ProtocolKey protocolKey

type ProtocolKeyPayload struct {
	TestRunID      uint        `json:"testRunId"`
	TestProtocolID uint        `json:"testProtocolId"`
	Key            interface{} `json:"key"`
}

type protocolKey struct {
	handlers []interface{ Handle(ProtocolKeyPayload) }
}

func (u *protocolKey) Register(handler interface{ Handle(ProtocolKeyPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u protocolKey) Trigger(payload ProtocolKeyPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
