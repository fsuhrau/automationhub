package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

// DeviceListChannel is the SSE channel on which the UI learns that it has to
// reload the device list (a device was added or its details changed).
const DeviceListChannel = "device_list"

type deviceListNotifier struct {
	publisher sse.Publisher
}

func RegisterEventDeviceListListener(publisher sse.Publisher) {
	events.DeviceListChanged.Register(deviceListNotifier{publisher: publisher})
}

func (u deviceListNotifier) Handle(payload events.DeviceListChangedPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: DeviceListChannel,
		Content: payload,
	})
}
