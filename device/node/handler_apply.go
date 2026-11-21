package node

import (
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
)

// nodeSync is one application of a node's device list; every device it
// touches is stamped with now.
type nodeSync struct {
	h          *Handler
	identifier manager.NodeIdentifier
	nodeID     uint
	now        time.Time
}

// apply updates the known device the node reported, or adds it when it is new.
func (s nodeSync) apply(devices []device.Device, reported device.Device) []device.Device {
	if known := findNodeDevice(devices, reported.DeviceID()); known != nil {
		s.update(known, reported)
		return devices
	}
	return append(devices, s.add(reported))
}

func findNodeDevice(devices []device.Device, id string) *NodeDevice {
	for _, d := range devices {
		if nd, ok := d.(*NodeDevice); ok && nd.DeviceID() == id {
			return nd
		}
	}
	return nil
}

// add makes a device new to the master known and stores it.
func (s nodeSync) add(reported device.Device) *NodeDevice {
	nd := s.h.newNodeDevice(s.identifier)
	nd.refresh(reported, s.now)
	model := models.Device{
		NodeID:           s.nodeID,
		DeviceIdentifier: nd.ID,
		DeviceType:       models.DeviceType(nd.Type),
		Name:             nd.Name,
		Manager:          Manager,
		OS:               nd.OSName,
		PlatformType:     models.PlatformType(nd.Platform),
		OSVersion:        nd.OSVersion,
		ConnectionType:   models.ConnectionTypeNode,
	}
	nd.SetConfig(&model)
	s.h.logStoreError("store new device", s.h.deviceStorage.NewDevice(s.h.Name(), model))
	s.h.store(nd)
	return nd
}

func (s nodeSync) update(nd *NodeDevice, reported device.Device) {
	nd.refresh(reported, s.now)
	if nd.GetConfig() == nil {
		nd.Config, _ = s.h.deviceStorage.GetDevice(Manager, nd.ID)
	}
	s.h.store(nd)
}

// markUnreported sets the state of the devices the node didn't report this
// time: NodeDisconnected when the node itself is offline, otherwise
// RemoteDisconnected.
func (s nodeSync) markUnreported(managers map[string][]device.Device, nodeOffline bool) {
	state := device.StateRemoteDisconnected
	if nodeOffline {
		state = device.StateNodeDisconnected
	}
	for _, devs := range managers {
		for _, dev := range devs {
			nd, ok := dev.(*NodeDevice)
			if ok && !nd.lastUpdateAt.Equal(s.now) {
				nd.State = state
				s.h.store(nd)
			}
		}
	}
}

// store saves the device state; a failure is logged, the sync goes on.
func (h *Handler) store(nd *NodeDevice) {
	h.logStoreError("update device", h.deviceStorage.Update(h.Name(), nd))
}

func (h *Handler) logStoreError(what string, err error) {
	if err != nil {
		h.log.Warnf("%s: %v", what, err)
	}
}
