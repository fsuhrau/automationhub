package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

// testRunChange is sent on the "test_runs" channel whenever a run started or
// finished, so pages that list runs (dashboard, tests, run history) refresh no
// matter whether the run was started from the UI, the CLI or MCP.
type testRunChange struct {
	ProjectID string `json:"projectId"`
	AppID     uint   `json:"appId"`
	TestID    uint   `json:"testId,omitempty"`
	TestRunID uint   `json:"testRunId"`
	State     string `json:"state"`
}

type testRunChangedNotifier struct {
	publisher sse.Publisher
}

func RegisterEventTestRunChangedListener(publisher sse.Publisher) {
	notifier := testRunChangedNotifier{publisher: publisher}
	events.TestRunStarted.Register(notifier)
	events.TestRunFinished.Register(notifier.finished())
}

func (u testRunChangedNotifier) Handle(payload events.TestRunStartedPayload) {
	u.publish(testRunChange{ProjectID: payload.ProjectID, AppID: payload.AppID, TestID: payload.TestID, TestRunID: payload.TestRunID, State: "started"})
}

func (u testRunChangedNotifier) publish(change testRunChange) {
	u.publisher.PublishEvent(sse.Event{Channel: "test_runs", Content: change})
}

type testRunFinishedChange struct{ testRunChangedNotifier }

func (u testRunChangedNotifier) finished() testRunFinishedChange {
	return testRunFinishedChange{u}
}

func (u testRunFinishedChange) Handle(payload events.TestRunFinishedPayload) {
	u.publish(testRunChange{ProjectID: payload.ProjectID, AppID: payload.AppID, TestRunID: payload.TestRunID, State: "finished"})
}
