package api

import (
	"fmt"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
)

type newTestProtocolNotifier struct {
	publisher sse.Publisher
}

func RegisterNewTestProtocolListener(publisher sse.Publisher) {
	notifier := newTestProtocolNotifier{
		publisher: publisher,
	}

	events.NewTestProtocol.Register(notifier)
}

func (u newTestProtocolNotifier) Handle(payload events.NewTestProtocolPayload) {
	content := payload.Protocol
	if p, ok := content.(*models.TestProtocol); ok {
		content = protocolForUI(p)
	}
	u.publisher.PublishEvent(sse.Event{
		Channel: fmt.Sprintf("test_run_%d_protocol", payload.TestRunID),
		Content: content,
	})
}

// protocolForUI copies the protocol without the live device driver and
// connection: a driver points back to its device config, so the protocol
// can't be encoded as JSON and the run page would never see the test.
func protocolForUI(p *models.TestProtocol) *models.TestProtocol {
	c := *p
	if p.Device != nil {
		d := *p.Device
		d.Dev = nil
		d.Connection = nil
		c.Device = &d
	}
	return &c
}
