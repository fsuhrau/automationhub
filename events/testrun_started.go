package events

// TestRunStarted is triggered when a test run was started, from the web UI,
// the CLI or MCP alike.
var TestRunStarted testRunStarted

type TestRunStartedPayload struct {
	ProjectID string `json:"projectId"`
	AppID     uint   `json:"appId"`
	TestID    uint   `json:"testId"`
	TestRunID uint   `json:"testRunId"`
}

type testRunStarted struct {
	handlers []interface{ Handle(TestRunStartedPayload) }
}

func (u *testRunStarted) Register(handler interface{ Handle(TestRunStartedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u testRunStarted) Trigger(payload TestRunStartedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
