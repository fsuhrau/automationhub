package events

var TestRunFinished testRunFinished

type TestRunFinishedPayload struct {
	ProjectID string      `json:"projectId"`
	AppID     uint        `json:"appId"`
	TestRunID uint        `json:"testRunId"`
	TestName  string      `json:"testName"`
	Success   bool        `json:"success"`
	Succeeded int         `json:"succeeded"`
	Unstable  int         `json:"unstable"`
	Failed    int         `json:"failed"`
	TestRun   interface{} `json:"testRun"`
}

type testRunFinished struct {
	handlers []interface{ Handle(TestRunFinishedPayload) }
}

func (u *testRunFinished) Register(handler interface{ Handle(TestRunFinishedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u testRunFinished) Trigger(payload TestRunFinishedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
