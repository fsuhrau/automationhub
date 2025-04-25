package events

var NewTestLogEntry newTestLogEntry

type NewTestLogEntryPayload struct {
	TestRunID uint        `json:"testRunId"`
	Entry     interface{} `json:"entry"`
}

type newTestLogEntry struct {
	handlers []interface{ Handle(NewTestLogEntryPayload) }
}

func (u *newTestLogEntry) Register(handler interface{ Handle(NewTestLogEntryPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u newTestLogEntry) Trigger(payload NewTestLogEntryPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
