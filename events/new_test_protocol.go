package events

var NewTestProtocol newTestProtocol

type NewTestProtocolPayload struct {
	TestRunID uint
	Protocol  interface{}
}

type newTestProtocol struct {
	handlers []interface{ Handle(NewTestProtocolPayload) }
}

func (u *newTestProtocol) Register(handler interface{ Handle(NewTestProtocolPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u newTestProtocol) Trigger(payload NewTestProtocolPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
