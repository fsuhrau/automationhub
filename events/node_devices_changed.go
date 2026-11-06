package events

import "github.com/fsuhrau/automationhub/hub/manager"

// NodeDevicesChanged is triggered when a node reports that its device list
// changed (device plugged in, unplugged, authorized, ...).
var NodeDevicesChanged nodeDevicesChanged

type NodeDevicesChangedPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
}

type nodeDevicesChanged struct {
	handlers []interface {
		Handle(NodeDevicesChangedPayload)
	}
}

func (u *nodeDevicesChanged) Register(handler interface {
	Handle(NodeDevicesChangedPayload)
}) {
	u.handlers = append(u.handlers, handler)
}

func (u nodeDevicesChanged) Trigger(payload NodeDevicesChangedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
