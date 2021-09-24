package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type testCreatedNotifier struct{
	publisher sse.Publisher
}

func RegisterEventTestCreatedListener(publisher sse.Publisher) {
	notifier := testCreatedNotifier{
		publisher: publisher,
	}

	events.TestCreated.Register(notifier)
}

func (u testCreatedNotifier) Handle(payload events.TestCreatedPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: "tests",
		Content: payload,
	})
}