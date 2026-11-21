package node

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Init registers the node events, restores the stored node devices and asks
// the nodes for their current ones. The arguments are the node settings of
// the other handlers; the master side needs none of them.
func (h *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	h.init = true
	defer func() {
		h.init = false
	}()

	h.registerListeners()
	if err := h.restoreStoredDevices(); err != nil {
		return err
	}
	return h.RefreshDevices(true)
}

func (h *Handler) registerListeners() {
	events.NodeConnected.Register(h)
	events.NodeDevicesChanged.Register(devicesChangedListener{h})
	events.NodeDisconnected.Register(nodeDisconnectedListener{h})
}

// restoreStoredDevices makes the node devices of the database known, state
// unknown until their node reports them.
func (h *Handler) restoreStoredDevices() error {
	stored, err := h.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for _, model := range stored {
		if err := h.restoreDevice(model); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) restoreDevice(model *models.Device) error {
	var n models.Node
	if err := h.db.First(&n, "id = ?", model.NodeID).Error; err != nil {
		return err
	}
	dev := h.newNodeDevice(manager.NodeIdentifier(n.Identifier))
	dev.OSName = model.OS
	dev.OSVersion = model.OSVersion
	dev.Target = model.TargetVersion
	dev.Name = model.Name
	dev.ID = model.DeviceIdentifier
	dev.Platform = int(model.PlatformType)
	dev.State = device.StateUnknown
	dev.SetConfig(model)
	h.addDevice(dev.nodeId, model.Manager, dev)
	return nil
}

// addDevice makes dev known under its node and the manager name.
func (h *Handler) addDevice(node manager.NodeIdentifier, managerName string, dev device.Device) {
	h.mu.Lock()
	defer h.mu.Unlock()
	managed := h.nodes[node]
	if managed == nil {
		managed = make(map[string][]device.Device)
		h.nodes[node] = managed
	}
	managed[managerName] = append(managed[managerName], dev)
}
