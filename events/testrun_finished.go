package events

var TestRunFinished testRunFinished

type TestRunFinishedPayload struct {
	ProjectID string
	AppID     uint
	TestRunID uint
	TestName  string
	Success   bool
	Succeeded int
	Unstable  int
	Failed    int
	TestRun   interface{}
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
