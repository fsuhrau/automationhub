package events

var NewTestLogEntry = NewBus[NewTestLogEntryPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NewTestLogEntryPayload struct {
	TestRunID uint        `json:"testRunId"`
	Entry     interface{} `json:"entry"`
}
