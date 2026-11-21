package iossim

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/storage"
)

const (
	Manager = "ios_sim"

	simulatorApp = "/Applications/Xcode.app/Contents/Developer/Applications/Simulator.app/Contents/MacOS/Simulator"
)

type Handler struct {
	generic.Registry[*Device]
	mu            sync.Mutex // serializes refreshes
	deviceStorage storage.Device
	deps          Deps
	simctl        simctl
}

func NewHandler(ds storage.Device) *Handler {
	return NewHandlerWithDeps(ds, Deps{})
}

func NewHandlerWithDeps(ds storage.Device, deps Deps) *Handler {
	deps = deps.withDefaults()
	return &Handler{deviceStorage: ds, deps: deps, simctl: simctl{runner: deps.Runner}}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterURL, nodeIdentifier string, authToken *string) error {
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return fmt.Errorf("load simulators: %w", err)
	}
	for _, cfg := range devs {
		dev := NewDevice(generic.Info{
			ID:        cfg.DeviceIdentifier,
			Name:      cfg.Name,
			OSName:    cfg.OS,
			OSVersion: cfg.OSVersion,
			State:     device.StateUnknown,
		}, m.deps)
		dev.SetConfig(cfg)
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterURL, nodeIdentifier, dev.ID, authToken))
		dev.AddActionHandler(remote.NewActionHandler(masterURL, nodeIdentifier, dev.ID, authToken))
		m.Add(dev)
	}
	return nil
}

func (m *Handler) Start() error {
	return m.command("open", "-a", simulatorApp)
}

func (m *Handler) Stop() error {
	return m.command("killall", "Simulator")
}

func (m *Handler) command(name string, args ...string) error {
	if _, err := m.deps.Runner.Run(context.Background(), name, args...); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func (m *Handler) StartDevice(deviceID string) error {
	return m.simctlKnown("boot", deviceID)
}

func (m *Handler) StopDevice(deviceID string) error {
	return m.simctlKnown("shutdown", deviceID)
}

// simctlKnown runs the simctl command for a simulator the handler manages.
func (m *Handler) simctlKnown(command, deviceID string) error {
	if !m.Has(deviceID) {
		return device.DeviceNotFoundError
	}
	_, err := m.simctl.run(command, deviceID)
	return err
}

func (m *Handler) RegisterDevice(device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device: %w", device.ErrUnsupported)
}
