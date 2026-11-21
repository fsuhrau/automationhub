package node

import (
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
)

// RefreshDevices fetches the device lists of all nodes over RPC. Nodes push
// their changes (UpdateNodeDevices), so this runs when the master starts, when
// a node connects or drops, and every UpdateInterval as a safety net.
func (h *Handler) RefreshDevices(force bool) error {
	h.refreshMu.Lock()
	defer h.refreshMu.Unlock()

	now := time.Now().UTC()
	if !force && h.lastUpdate.Add(UpdateInterval).After(now) {
		return nil
	}
	h.lastUpdate = now

	for _, nodeIdentifier := range h.knownNodes() {
		if err := h.syncNode(nodeIdentifier, h.fetchNodeDevices(nodeIdentifier), now); err != nil {
			return err
		}
	}
	return nil
}

// knownNodes returns the connected nodes and the nodes with known devices.
func (h *Handler) knownNodes() []manager.NodeIdentifier {
	connected := h.nodeManager.GetNodes()

	h.mu.Lock()
	defer h.mu.Unlock()
	for _, n := range connected {
		if _, ok := h.nodes[n]; !ok {
			h.nodes[n] = make(map[string][]device.Device)
		}
	}
	known := make([]manager.NodeIdentifier, 0, len(h.nodes))
	for n := range h.nodes {
		known = append(known, n)
	}
	return known
}

// refreshNode fetches the device list of one node over RPC (a node that
// announced a change without sending its list).
func (h *Handler) refreshNode(nodeIdentifier manager.NodeIdentifier) error {
	h.refreshMu.Lock()
	defer h.refreshMu.Unlock()
	return h.syncNode(nodeIdentifier, h.fetchNodeDevices(nodeIdentifier), time.Now().UTC())
}

// UpdateNodeDevices applies a device list a node pushed. Pushes older than
// the last applied one (sentAt, node clock) are dropped, and so are pushes of
// a node that isn't connected: it is refreshed completely when it connects.
func (h *Handler) UpdateNodeDevices(nodeIdentifier manager.NodeIdentifier, managers map[string][]device.Device, sentAt int64) error {
	h.refreshMu.Lock()
	defer h.refreshMu.Unlock()

	if status, _ := h.nodeManager.GetStatus(nodeIdentifier); status == 0 {
		return nil
	}
	if sentAt != 0 {
		if sentAt <= h.lastPush[nodeIdentifier] {
			return nil
		}
		h.lastPush[nodeIdentifier] = sentAt
	}
	if managers == nil {
		managers = make(map[string][]device.Device)
	}
	return h.syncNode(nodeIdentifier, managers, time.Now().UTC())
}

// fetchNodeDevices asks a node for its devices; nil when the node is not
// connected or doesn't answer.
func (h *Handler) fetchNodeDevices(nodeIdentifier manager.NodeIdentifier) map[string][]device.Device {
	if status, _ := h.nodeManager.GetStatus(nodeIdentifier); status == 0 {
		return nil
	}
	managers, err := h.nodeManager.GetManagers(nodeIdentifier)
	if err != nil {
		return nil
	}
	if managers == nil {
		managers = make(map[string][]device.Device)
	}
	return managers
}

// syncNode applies the device list of one node. nodeDevices nil means the
// node is offline: all its devices become NodeDisconnected. Otherwise devices
// the node no longer reports (unplugged, stopped) become RemoteDisconnected.
// The caller holds refreshMu.
func (h *Handler) syncNode(nodeIdentifier manager.NodeIdentifier, nodeDevices map[string][]device.Device, now time.Time) error {
	var record models.Node
	if err := h.db.First(&record, "identifier = ?", string(nodeIdentifier)).Error; err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	managers := h.nodes[nodeIdentifier]
	if managers == nil {
		managers = make(map[string][]device.Device)
		h.nodes[nodeIdentifier] = managers
	}
	run := nodeSync{h: h, identifier: nodeIdentifier, nodeID: record.ID, now: now}
	devices := managers[Manager]
	for _, reported := range nodeDevices {
		for _, d := range reported {
			devices = run.apply(devices, d)
		}
	}
	managers[Manager] = devices

	run.markUnreported(managers, nodeDevices == nil)
	return nil
}
