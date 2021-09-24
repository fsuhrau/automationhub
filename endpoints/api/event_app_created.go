package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type appCreatedNotifier struct{
	publisher sse.Publisher
}

func RegisterEventAppCreatedListener(publisher sse.Publisher) {
	notifier := appCreatedNotifier{
		publisher: publisher,
	}

	events.AppCreated.Register(notifier)
}

func (u appCreatedNotifier) Handle(payload events.AppCreatedPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: "apps",
		Content: payload,
	})
}