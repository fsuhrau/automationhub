package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type newTestProtocolNotifier struct{
	publisher sse.Publisher
}

func RegisterNewTestProtocolListener(publisher sse.Publisher) {
	notifier := newTestProtocolNotifier{
		publisher: publisher,
	}

	events.NewTestProtocol.Register(notifier)
}

func (u newTestProtocolNotifier) Handle(payload events.NewTestProtocolPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_run_%d_protocol", payload.TestRunID),
		Content: payload,
	})
}