package events

var TestRunFinished = NewBus[TestRunFinishedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

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
