package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
)

type protocolKeyNotifier struct {
	publisher sse.Publisher
}

func RegisterProtocolKeyListener(publisher sse.Publisher) {
	events.ProtocolKey.Register(protocolKeyNotifier{publisher: publisher})
}

// Handle tells the protocol page about the key.
func (u protocolKeyNotifier) Handle(payload events.ProtocolKeyPayload) {
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_protocol_%d_keys", payload.TestProtocolID),
		Content: payload,
	})
}
