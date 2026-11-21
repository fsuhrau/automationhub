package unityeditor

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
)

const (
	Manager = "unity_editor"
	// staleAfter is how long a registered editor may stay silent.
	staleAfter = time.Minute
)

// Deps are the external collaborators of the Handler. Zero fields get the real
// implementations.
type Deps struct {
	Runner device.CommandRunner
	Clock  device.Clock
	// GOOS is the operating system the editor runs on, runtime.GOOS by default.
	GOOS string
}

type Handler struct {
	generic.Registry[*Device]

	deviceStorage storage.Device
	managerCfg    config.Manager
	runner        device.CommandRunner
	clock         device.Clock
	goos          string
	// mu makes registering a device atomic.
	mu sync.Mutex
}

func NewHandler(cfg config.Manager, ds storage.Device) *Handler {
	return NewHandlerWithDeps(cfg, ds, Deps{})
}

func NewHandlerWithDeps(cfg config.Manager, ds storage.Device, deps Deps) *Handler {
	deps = deps.withDefaults()
	return &Handler{deviceStorage: ds, managerCfg: cfg, runner: deps.Runner, clock: deps.Clock, goos: deps.GOOS}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return fmt.Errorf("load unity editors: %w", err)
	}
	for i := range devs {
		dev := newDevice(generic.Info{
			ID:        devs[i].DeviceIdentifier,
			Name:      devs[i].Name,
			OSName:    devs[i].OS,
			OSVersion: devs[i].OSVersion,
			Target:    devs[i].TargetVersion,
		}, m.clock)
		dev.osInfos = devs[i].OSInfos
		dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.DeviceID(), authToken))
		dev.AddActionHandler(remote.NewActionHandler(masterUrl, nodeIdentifier, dev.DeviceID(), authToken))
		m.Add(dev)
	}
	return m.RefreshDevices(true)
}

func (m *Handler) Start() error {
	return nil
}

func (m *Handler) Stop() error {
	return nil
}

func (m *Handler) StopDevice(deviceID string) error {
	dev, ok := m.Get(deviceID)
	if !ok {
		return nil
	}
	if err := dev.stopProcess(); err != nil {
		return err
	}
	dev.UnityLogStopListening()
	return nil
}

// RefreshDevices marks editors that were silent for too long as shut down and
// stores the devices that changed.
func (m *Handler) RefreshDevices(_ bool) error {
	now := m.clock.Now().UTC()
	for _, dev := range m.List() {
		if dev.expire(now) || dev.takeUpdated() {
			m.storeUpdate(dev)
		}
	}
	return nil
}

func (m *Handler) storeUpdate(dev *Device) {
	if err := m.deviceStorage.Update(Manager, dev); err != nil {
		logrus.WithError(err).Warnf("unity editor %s: store update", dev.DeviceID())
	}
}

// expire shuts the device down when it was silent since staleAfter and
// reports whether its state changed.
func (d *Device) expire(now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if now.Sub(d.lastUpdateAt) <= staleAfter || d.State == device.StateShutdown {
		return false
	}
	d.State = device.StateShutdown
	d.updated = false
	return true
}

// takeUpdated reports and clears the updated flag.
func (d *Device) takeUpdated() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	updated := d.updated
	d.updated = false
	return updated
}
