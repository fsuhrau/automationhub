package events

var TestCreated testCreated

type TestCreatedPayload struct {
	TestID uint
	Test   interface{}
}

type testCreated struct {
	handlers []interface{ Handle(TestCreatedPayload) }
}

func (u *testCreated) Register(handler interface{ Handle(TestCreatedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u testCreated) Trigger(payload TestCreatedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
