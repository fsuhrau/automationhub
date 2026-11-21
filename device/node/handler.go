package node

import (
	"context"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	Manager = "node"
	// UpdateInterval is the safety net: nodes push every change of their
	// device list, so the master only asks them this often.
	UpdateInterval = 5 * time.Minute
)

// Handler manages the devices of all nodes on the master.
type Handler struct {
	log           *logrus.Entry
	mu            sync.RWMutex                     // guards nodes
	refreshMu     sync.Mutex                       // serializes RefreshDevices and node pushes
	lastPush      map[manager.NodeIdentifier]int64 // sentAt of the last applied push per node
	nodes         map[manager.NodeIdentifier]map[string][]device.Device
	deviceStorage storage.Device
	nodeManager   manager.Nodes
	rendezvous    *videostream.Rendezvous // pairs the video streams of the node devices
	init          bool
	db            *gorm.DB
	lastUpdate    time.Time
}

// NewHandler returns the handler of the node devices. The video streams of
// its devices pair through the rendezvous of nm when nm provides one
// (videostream.Provider), through a private one otherwise.
//
//nolint:revive // argument-limit: the signature is fixed by its caller in hub/service_master.go
func NewHandler(log *logrus.Logger, ds storage.Device, nm manager.Nodes, db *gorm.DB) *Handler {
	return &Handler{
		log:           log.WithField("Handler", "Node"),
		nodes:         make(map[manager.NodeIdentifier]map[string][]device.Device),
		lastPush:      make(map[manager.NodeIdentifier]int64),
		nodeManager:   nm,
		rendezvous:    videostream.From(nm),
		deviceStorage: ds,
		db:            db,
	}
}

func (h *Handler) Run(ctx context.Context, runSocketListener bool) error {
	nodes := h.nodeManager.GetNodes()
	for _, n := range nodes {
		managers, err := h.nodeManager.GetManagers(n)
		if err != nil {
			continue
		}

		h.mu.Lock()
		h.nodes[n] = managers
		h.mu.Unlock()
	}
	return nil
}

func (h *Handler) Name() string {
	return Manager
}

func (h *Handler) Handle(payload events.NodeConnectedPayload) {
	h.forgetPushes(payload.NodeIdentifier)
	_ = h.RefreshDevices(true)
}

// forgetPushes resets the push order of a node that (re)connected or dropped,
// so a node whose clock went back isn't ignored afterwards.
func (h *Handler) forgetPushes(nodeIdentifier manager.NodeIdentifier) {
	h.refreshMu.Lock()
	delete(h.lastPush, nodeIdentifier)
	h.refreshMu.Unlock()
}

// devicesChangedListener applies a device list a node pushed (the periodic
// refresh is only a safety net).
type devicesChangedListener struct{ h *Handler }

func (l devicesChangedListener) Handle(payload events.NodeDevicesChangedPayload) {
	var err error
	if payload.Devices != nil {
		err = l.h.UpdateNodeDevices(payload.NodeIdentifier, payload.Devices, payload.SentAt)
	} else {
		err = l.h.refreshNode(payload.NodeIdentifier)
	}
	if err != nil {
		l.h.log.Warnf("device update of node %s: %v", payload.NodeIdentifier, err)
	}
}

// nodeDisconnectedListener marks the devices of a lost node right away.
type nodeDisconnectedListener struct{ h *Handler }

func (l nodeDisconnectedListener) Handle(payload events.NodeDisconnectedPayload) {
	l.h.forgetPushes(payload.NodeIdentifier)
	_ = l.h.RefreshDevices(true)
}
