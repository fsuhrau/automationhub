package api

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type eventRecorder chan sse.Event

func (r eventRecorder) PublishEvent(e sse.Event) { r <- e }

// A run started or finished anywhere (UI, CLI, MCP) must reach the pages that
// list runs through the "test_runs" channel.
func TestTestRunStartAndFinishArePublished(t *testing.T) {
	r := make(eventRecorder, 8)
	RegisterEventTestRunChangedListener(r)

	events.TestRunStarted.Trigger(events.TestRunStartedPayload{ProjectID: "p", AppID: 2, TestID: 3, TestRunID: 41})
	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{ProjectID: "p", AppID: 2, TestRunID: 41})

	seen := map[string]testRunChange{}
	deadline := time.After(time.Second)
	for len(seen) < 2 {
		select {
		case e := <-r:
			if e.Channel != "test_runs" {
				t.Fatalf("event on channel %q, want test_runs", e.Channel)
			}
			c := e.Content.(testRunChange)
			if c.TestRunID == 41 {
				seen[c.State] = c
			}
		case <-deadline:
			t.Fatalf("missing test_runs events, got %v", seen)
		}
	}
	if s := seen["started"]; s.ProjectID != "p" || s.AppID != 2 || s.TestID != 3 {
		t.Errorf("started event %+v", s)
	}
	if f, ok := seen["finished"]; !ok || f.AppID != 2 {
		t.Errorf("finished event %+v", f)
	}
}
