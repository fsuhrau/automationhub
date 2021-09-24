package api

import (
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
		Channel: "testlog",
		Content: payload,
	})
}