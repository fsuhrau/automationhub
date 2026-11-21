package events

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// NodeDevicesChanged is triggered when a node reports that its device list
// changed (device plugged in, unplugged, authorized, ...).
var NodeDevicesChanged = NewBus[NodeDevicesChangedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type NodeDevicesChangedPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
	// Devices is the node's complete device list keyed by manager. nil when
	// the node didn't send it (older nodes); then it is fetched over RPC.
	Devices map[string][]device.Device `json:"-"`
	// SentAt orders pushes of one node (unix nano, node clock).
	SentAt int64 `json:"sentAt"`
}
