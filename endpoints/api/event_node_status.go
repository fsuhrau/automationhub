package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
)

// NodeStatusPayload is published on the "nodes" SSE channel.
type NodeStatusPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
	Connected      bool                   `json:"connected"`
}

type nodeStatusNotifier struct {
	publisher sse.Publisher
}

func RegisterNodeStatusListener(publisher sse.Publisher) {
	n := nodeStatusNotifier{publisher: publisher}
	events.NodeConnected.Register(n)
	events.NodeDisconnected.Register(nodeDisconnectedNotifier{n})
}

func (n nodeStatusNotifier) Handle(payload events.NodeConnectedPayload) {
	n.publisher.PublishEvent(sse.Event{Channel: "nodes", Content: NodeStatusPayload{NodeIdentifier: payload.NodeIdentifier, Connected: true}})
}

type nodeDisconnectedNotifier struct{ n nodeStatusNotifier }

func (d nodeDisconnectedNotifier) Handle(payload events.NodeDisconnectedPayload) {
	d.n.publisher.PublishEvent(sse.Event{Channel: "nodes", Content: NodeStatusPayload{NodeIdentifier: payload.NodeIdentifier, Connected: false}})
}
