package events

var NewTestProtocolLog = NewBus[NewTestProtocolLogPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NewTestProtocolLogPayload struct {
	TestProtocolID uint        `json:"testProtocolId"`
	Entry          interface{} `json:"entry"`
}
