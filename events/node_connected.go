package events

import "github.com/fsuhrau/automationhub/hub/manager"

var NodeConnected = NewBus[NodeConnectedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NodeConnectedPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
	Hostname       string                 `json:"hostname"`
}
