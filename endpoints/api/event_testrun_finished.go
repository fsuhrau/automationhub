package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type testRunFinishedNotifier struct{
	publisher sse.Publisher
}

func RegisterEventTestRunFinishedListener(publisher sse.Publisher) {
	notifier := testRunFinishedNotifier{
		publisher: publisher,
	}
	events.TestRunFinished.Register(notifier)
}

func (u testRunFinishedNotifier) Handle(payload events.TestRunFinishedPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_run_%d_finished", payload.TestRunID),
		Content: payload,
	})
}