package events

import "github.com/fsuhrau/automationhub/hub/manager"

var NodeConnected nodeConnected

type NodeConnectedPayload struct {
	NodeIdentifier manager.NodeIdentifier
	Hostname       string
}

type nodeConnected struct {
	handlers []interface{ Handle(NodeConnectedPayload) }
}

func (u *nodeConnected) Register(handler interface{ Handle(NodeConnectedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u nodeConnected) Trigger(payload NodeConnectedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
