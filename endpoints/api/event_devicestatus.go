package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type deviceStatusNotifier struct{
	publisher sse.Publisher
}

func RegisterEventDeviceStatusListener(publisher sse.Publisher) {
	notifier := deviceStatusNotifier{
		publisher: publisher,
	}

	events.DeviceStatusChanged.Register(notifier)
}

func (u deviceStatusNotifier) Handle(payload events.DeviceStatusChangedPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: "devices",
		Content: payload,
	})
}