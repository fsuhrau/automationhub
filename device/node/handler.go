package node

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/device/unityeditor"
	"github.com/fsuhrau/automationhub/device/web"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	Manager = "node"
	// UpdateInterval is the safety net: nodes push every change of their
	// device list, so the master only asks them this often.
	UpdateInterval = 5 * time.Minute
)

type Handler struct {
	log           *logrus.Entry
	mu            sync.RWMutex                     // guards nodes
	refreshMu     sync.Mutex                       // serializes RefreshDevices and node pushes
	lastPush      map[manager.NodeIdentifier]int64 // sentAt of the last applied push per node
	nodes         map[manager.NodeIdentifier]map[string][]device.Device
	deviceStorage storage.Device
	nodeManager   manager.Nodes
	init          bool
	db            *gorm.DB
	//devices       map[string]*NodeDevice
	lastUpdate time.Time
}

func NewHandler(log *logrus.Logger, ds storage.Device, nm manager.Nodes, db *gorm.DB) *Handler {
	return &Handler{
		log:      log.WithField("Handler", "Node"),
		nodes:    make(map[manager.NodeIdentifier]map[string][]device.Device),
		lastPush: make(map[manager.NodeIdentifier]int64),
		//devices:       make(map[string]*NodeDevice),
		nodeManager:   nm,
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

func (h *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	h.init = true
	defer func() {
		h.init = false
	}()

	events.NodeConnected.Register(h)
	events.NodeDevicesChanged.Register(devicesChangedListener{h})
	events.NodeDisconnected.Register(nodeDisconnectedListener{h})

	devs, err := h.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for i := range devs {
		var n models.Node
		if err := h.db.First(&n, "id = ?", devs[i].NodeID).Error; err != nil {
			return err
		}

		dev := &NodeDevice{
			nodeId:          manager.NodeIdentifier(n.Identifier),
			nodeManager:     h.nodeManager,
			deviceOSName:    devs[i].OS,
			deviceOSVersion: devs[i].OSVersion,
			targetVersion:   devs[i].TargetVersion,
			deviceName:      devs[i].Name,
			deviceID:        devs[i].DeviceIdentifier,
			platformType:    devs[i].PlatformType,
			deviceState:     device.StateUnknown,
		}
		dev.SetConfig(devs[i])

		h.mu.Lock()
		managed := h.nodes[dev.nodeId]
		if managed == nil {
			managed = make(map[string][]device.Device)
		}
		devices := managed[devs[i].Manager]
		devices = append(devices, dev)
		managed[devs[i].Manager] = devices
		h.nodes[dev.nodeId] = managed
		h.mu.Unlock()
	}
	if err := h.RefreshDevices(true); err != nil {
		return err
	}

	return nil
}

func managerToPlatformType(mng string) models.PlatformType {
	if mng == androiddevice.Manager {
		return models.PlatformTypeAndroid
	}

	if mng == iosdevice.Manager {
		return models.PlatformTypeiOS
	}

	if mng == iossim.Manager {
		return models.PlatformTypeiOSSimulator
	}

	if mng == macos.Manager {
		return models.PlatformTypeMac
	}

	if mng == unityeditor.Manager {
		return models.PlatformTypeEditor
	}

	if mng == web.Manager {
		return models.PlatformTypeWeb
	}

	panic("manager not handled: " + mng)
}

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

	nodes := h.nodeManager.GetNodes()

	h.mu.Lock()
	for _, n := range nodes {
		if _, ok := h.nodes[n]; !ok {
			h.nodes[n] = make(map[string][]device.Device)
		}
	}
	knownNodes := make([]manager.NodeIdentifier, 0, len(h.nodes))
	for n := range h.nodes {
		knownNodes = append(knownNodes, n)
	}
	h.mu.Unlock()

	for _, nodeIdentifier := range knownNodes {
		if err := h.syncNode(nodeIdentifier, h.fetchNodeDevices(nodeIdentifier), now); err != nil {
			return err
		}
	}
	return nil
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
	var no models.Node
	if err := h.db.First(&no, "identifier = ?", string(nodeIdentifier)).Error; err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	managers := h.nodes[nodeIdentifier]
	if managers == nil {
		managers = make(map[string][]device.Device)
	}
	devices := managers[Manager]

	for _, devs := range nodeDevices {
		for _, d := range devs {
			deviceID := d.DeviceID()

			index := -1
			for i, d1 := range devices {
				if d1.DeviceID() == deviceID {
					index = i
					break
				}
			}
			if index < 0 {
				nd := &NodeDevice{
					nodeId:       nodeIdentifier,
					nodeManager:  h.nodeManager,
					deviceName:   d.DeviceName(),
					deviceID:     deviceID,
					deviceOSName: d.DeviceOSName(),
					platformType: models.PlatformType(d.PlatformType()),
					lastUpdateAt: now,
				}
				nd.UpdateDeviceInfos(d.DeviceParameter())
				nd.deviceState = d.DeviceState()
				dev := models.Device{
					NodeID:           no.ID,
					DeviceIdentifier: deviceID,
					DeviceType:       models.DeviceType(d.DeviceType()),
					Name:             d.DeviceName(),
					Manager:          Manager,
					OS:               d.DeviceOSName(),
					PlatformType:     models.PlatformType(d.PlatformType()),
					OSVersion:        d.DeviceOSVersion(),
					ConnectionType:   models.ConnectionTypeNode,
				}
				nd.SetConfig(&dev)
				devices = append(devices, nd)
				h.deviceStorage.NewDevice(h.Name(), dev)
				h.deviceStorage.Update(h.Name(), nd)
			} else {
				nd := devices[index].(*NodeDevice)
				nd.deviceType = d.DeviceType()
				nd.deviceOSName = d.DeviceOSName()
				nd.deviceOSVersion = d.DeviceOSVersion()
				nd.deviceName = d.DeviceName()
				nd.deviceID = deviceID
				nd.lastUpdateAt = now
				nd.deviceState = d.DeviceState()
				nd.platformType = models.PlatformType(d.PlatformType())
				nd.UpdateDeviceInfos(d.DeviceParameter())
				if nd.GetConfig() == nil {
					config, _ := h.deviceStorage.GetDevice(Manager, deviceID)
					nd.Config = config
				}
				devices[index] = nd
				h.deviceStorage.Update(h.Name(), nd)
			}
		}
	}
	managers[Manager] = devices
	h.nodes[nodeIdentifier] = managers

	state := device.StateRemoteDisconnected
	if nodeDevices == nil {
		state = device.StateNodeDisconnected
	}
	for _, devs := range managers {
		for _, dev := range devs {
			nd, ok := dev.(*NodeDevice)
			if ok && !nd.lastUpdateAt.Equal(now) {
				nd.deviceState = state
				h.deviceStorage.Update(h.Name(), nd)
			}
		}
	}
	return nil
}

func (h *Handler) Start() error {
	/* not needed */
	return nil
}

func (h *Handler) Stop() error {
	/* not needed */
	return nil
}

func (h *Handler) GetDevice(deviceID string) (error, manager.NodeIdentifier, device.Device) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for node, mng := range h.nodes {
		for _, devices := range mng {
			for _, dev := range devices {
				if dev.DeviceID() == deviceID {
					return nil, node, dev
				}
			}
		}
	}

	return device.DeviceNotFoundError, "", nil
}

func (h *Handler) StartDevice(deviceID string) error {
	err, node, _ := h.GetDevice(deviceID)
	if err != nil {
		return err
	}

	return h.nodeManager.StartDevice(node, deviceID)
}

func (h *Handler) StopDevice(deviceID string) error {
	err, node, _ := h.GetDevice(deviceID)
	if err != nil {
		return err
	}

	return h.nodeManager.StopDevice(node, deviceID)
}

func (h *Handler) GetDevices() ([]device.Device, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var devices []device.Device
	for _, mng := range h.nodes {
		for _, devs := range mng {
			devices = append(devices, devs...)
		}
	}
	return devices, nil
}

func (h *Handler) HasDevice(device2 device.Device) bool {
	err, _, _ := h.GetDevice(device2.DeviceID())
	return err == nil
}

func (h *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device not implemented")
}
