package events

import "github.com/fsuhrau/automationhub/hub/manager"

// NodeDisconnected is triggered when the master loses the connection to a node.
var NodeDisconnected nodeDisconnected

type NodeDisconnectedPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
}

type nodeDisconnected struct {
	handlers []interface{ Handle(NodeDisconnectedPayload) }
}

func (u *nodeDisconnected) Register(handler interface{ Handle(NodeDisconnectedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u nodeDisconnected) Trigger(payload NodeDisconnectedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
