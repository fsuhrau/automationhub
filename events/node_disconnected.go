package events

import "github.com/fsuhrau/automationhub/hub/manager"

// NodeDisconnected is triggered when the master loses the connection to a node.
var NodeDisconnected = NewBus[NodeDisconnectedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NodeDisconnectedPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
}
