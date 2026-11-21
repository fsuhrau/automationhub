package node

import (
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/videostream"
)

const ConnectionTimeout = 10 * time.Second

// NodeDevice is a device that lives on a node. It reports the identity the
// node announced and forwards everything else to the node through the node
// manager.
type NodeDevice struct {
	generic.Device
	generic.Info
	lastUpdateAt time.Time

	nodeId      manager.NodeIdentifier
	nodeManager manager.Nodes
	rendezvous  *videostream.Rendezvous
}

// newNodeDevice returns an empty device of the node; the caller fills in its
// identity.
func (h *Handler) newNodeDevice(node manager.NodeIdentifier) *NodeDevice {
	return &NodeDevice{
		Info:        generic.Info{Parameter: map[string]string{}},
		nodeId:      node,
		nodeManager: h.nodeManager,
		rendezvous:  h.rendezvous,
	}
}

// refresh takes over the identity and state the node reports for the device.
func (d *NodeDevice) refresh(reported device.Device, now time.Time) {
	d.ID = reported.DeviceID()
	d.Name = reported.DeviceName()
	d.OSName = reported.DeviceOSName()
	d.OSVersion = reported.DeviceOSVersion()
	d.Type = reported.DeviceType()
	d.Platform = reported.PlatformType()
	d.State = reported.DeviceState()
	d.UpdateDeviceInfos(reported.DeviceParameter())
	d.lastUpdateAt = now
}

func (d *NodeDevice) GetNodeID() manager.NodeIdentifier {
	return d.nodeId
}

func (d *NodeDevice) NodeManager() manager.Nodes {
	return d.nodeManager
}

func (d *NodeDevice) SetDeviceState(state string) {
	d.State = device.ParseState(state)
}

// UpdateDeviceInfos replaces the device parameters with a copy of parameter.
func (d *NodeDevice) UpdateDeviceInfos(parameter map[string]string) {
	d.Parameter = make(map[string]string, len(parameter))
	for k, v := range parameter {
		d.Parameter[k] = v
	}
}
