package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type newTestLogEntryNotifier struct{
	publisher sse.Publisher
}

func RegisterNewTestLogEntryListener(publisher sse.Publisher) {
	notifier := newTestLogEntryNotifier{
		publisher: publisher,
	}

	events.NewTestLogEntry.Register(notifier)
}

func (u newTestLogEntryNotifier) Handle(payload events.NewTestLogEntryPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_run_%d_log", payload.TestRunID),
		Content: payload,
	})
}