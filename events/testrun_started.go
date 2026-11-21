package events

// TestRunStarted is triggered when a test run was started, from the web UI,
// the CLI or MCP alike.
var TestRunStarted = NewBus[TestRunStartedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type TestRunStartedPayload struct {
	ProjectID string `json:"projectId"`
	AppID     uint   `json:"appId"`
	TestID    uint   `json:"testId"`
	TestRunID uint   `json:"testRunId"`
}
