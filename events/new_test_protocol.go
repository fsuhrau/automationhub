package events

var NewTestProtocol = NewBus[NewTestProtocolPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NewTestProtocolPayload struct {
	TestRunID uint        `json:"testRunId"`
	Protocol  interface{} `json:"protocol"`
}
