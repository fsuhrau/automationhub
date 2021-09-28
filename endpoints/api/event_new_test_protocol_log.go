package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type newTestProtocolLogNotifier struct{
	publisher sse.Publisher
}

func RegisterNewTestProtocolLogListener(publisher sse.Publisher) {
	notifier := newTestProtocolLogNotifier{
		publisher: publisher,
	}

	events.NewTestProtocolLog.Register(notifier)
}

func (u newTestProtocolLogNotifier) Handle(payload events.NewTestProtocolLogPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_protocol_%d_log", payload.TestProtocolID),
		Content: payload,
	})
}