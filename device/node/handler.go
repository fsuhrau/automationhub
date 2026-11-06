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
	Manager        = "node"
	UpdateInterval = 30 * time.Second
)

type Handler struct {
	log           *logrus.Entry
	mu            sync.RWMutex // guards nodes
	refreshMu     sync.Mutex   // serializes RefreshDevices
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
		log:   log.WithField("Handler", "Node"),
		nodes: make(map[manager.NodeIdentifier]map[string][]device.Device),
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
	_ = h.RefreshDevices(true)
}

// devicesChangedListener refreshes immediately when a node pushes a device
// change (the periodic refresh is only a fallback).
type devicesChangedListener struct{ h *Handler }

func (l devicesChangedListener) Handle(events.NodeDevicesChangedPayload) {
	_ = l.h.RefreshDevices(true)
}

func (h *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	h.init = true
	defer func() {
		h.init = false
	}()

	events.NodeConnected.Register(h)
	events.NodeDevicesChanged.Register(devicesChangedListener{h})

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

func (h *Handler) RefreshDevices(force bool) error {
	h.refreshMu.Lock()
	defer h.refreshMu.Unlock()

	lastUpdate := time.Now().UTC()
	if !force && h.lastUpdate.Add(UpdateInterval).After(lastUpdate) {
		return nil
	}
	h.lastUpdate = lastUpdate

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

	// nodes whose device list was fetched in this refresh
	refreshed := make(map[manager.NodeIdentifier]bool)

	for _, nodeIdentifier := range knownNodes {
		var no models.Node
		if err := h.db.First(&no, "identifier = ?", string(nodeIdentifier)).Error; err != nil {
			return err
		}

		nodeStatus, _ := h.nodeManager.GetStatus(nodeIdentifier)
		if nodeStatus == 0 {
			continue
		}

		nodeManagerManagers, err := h.nodeManager.GetManagers(nodeIdentifier)
		if err != nil {
			//h.log.Errorf("unable to get manager: %v", err)
			continue
		}

		refreshed[nodeIdentifier] = true

		// remote calls above are done without holding the lock
		h.mu.Lock()
		managers := h.nodes[nodeIdentifier]
		if managers == nil {
			managers = make(map[string][]device.Device)
		}
		var devices []device.Device
		if d1, ok := managers[Manager]; ok {
			devices = d1
		}

		for _, devs := range nodeManagerManagers {

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
						lastUpdateAt: lastUpdate,
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
					nd.lastUpdateAt = lastUpdate
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
		h.mu.Unlock()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	// devices missing from this refresh: either their node is offline, or the
	// node is connected but no longer reports the device (unplugged, stopped)
	for nodeIdentifier, managers := range h.nodes {
		state := device.StateNodeDisconnected
		if refreshed[nodeIdentifier] {
			state = device.StateRemoteDisconnected
		}
		for _, devices := range managers {
			for _, dev := range devices {
				nd, ok := dev.(*NodeDevice)
				if ok && nd.lastUpdateAt != lastUpdate {
					nd.deviceState = state
					h.deviceStorage.Update(h.Name(), nd)
				}
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
