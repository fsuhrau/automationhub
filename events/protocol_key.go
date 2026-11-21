package events

var ProtocolKey = NewBus[ProtocolKeyPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type ProtocolKeyPayload struct {
	TestRunID      uint        `json:"testRunId"`
	TestProtocolID uint        `json:"testProtocolId"`
	Key            interface{} `json:"key"`
}
