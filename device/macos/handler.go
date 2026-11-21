package macos

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/storage"
)

const Manager = "macos"

// Handler manages the one device of the Mac it runs on.
type Handler struct {
	generic.Registry[*Device]
	deviceStorage storage.Device
	deps          Deps
}

var _ device.Handler = (*Handler)(nil)

// NewHandler returns a Handler with the production dependencies.
func NewHandler(ds storage.Device) *Handler {
	return NewHandlerWith(ds, Deps{})
}

// NewHandlerWith returns a Handler with the given dependencies; unset ones get their defaults.
func NewHandlerWith(ds storage.Device, deps Deps) *Handler {
	return &Handler{deviceStorage: ds, deps: deps.withDefaults()}
}

func (m *Handler) Name() string { return Manager }

// Init creates the directories and registers this Mac as a device.
func (m *Handler) Init(masterURL, nodeIdentifier string, authToken *string) error {
	if err := m.deps.ensureDirs(); err != nil {
		return err
	}
	dev := newDevice(context.Background(), m.deps)
	dev.SetLogWriter(generic.NewRemoteLogWriter(masterURL, nodeIdentifier, dev.DeviceID(), authToken))
	dev.AddActionHandler(remote.NewActionHandler(masterURL, nodeIdentifier, dev.DeviceID(), authToken))
	m.Add(dev)
	if err := m.deviceStorage.Update(m.Name(), dev); err != nil {
		return fmt.Errorf("store device: %w", err)
	}
	return m.RefreshDevices(true)
}

func (m *Handler) Start() error { return nil }

func (m *Handler) Stop() error { return nil }

func (m *Handler) StartDevice(string) error {
	return fmt.Errorf("start device: %w", device.ErrUnsupported)
}

func (m *Handler) StopDevice(string) error {
	return fmt.Errorf("stop device: %w", device.ErrUnsupported)
}

// RefreshDevices does nothing, the Mac is registered once by Init.
func (m *Handler) RefreshDevices(bool) error { return nil }

func (m *Handler) RegisterDevice(device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device: %w", device.ErrUnsupported)
}
