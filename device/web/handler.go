package web

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/storage"
)

const Manager = "web"

// Handler manages one device per configured browser.
type Handler struct {
	generic.Registry[*Device]
	cfg           config.Manager
	deviceStorage storage.Device
	env           hostinfo.Env
}

var _ device.Handler = (*Handler)(nil)

// NewHandler returns a Handler with the production dependencies.
func NewHandler(cfg config.Manager, ds storage.Device) *Handler {
	return NewHandlerWith(cfg, ds, hostinfo.Env{})
}

// NewHandlerWith returns a Handler with the given environment; unset fields get their defaults.
func NewHandlerWith(cfg config.Manager, ds storage.Device, env hostinfo.Env) *Handler {
	return &Handler{cfg: cfg, deviceStorage: ds, env: env.WithDefaults()}
}

func (m *Handler) Name() string { return Manager }

// Init registers a device for every configured browser.
func (m *Handler) Init(masterURL, nodeIdentifier string, authToken *string) error {
	for browser, path := range m.cfg.Browser {
		dev := newDevice(context.Background(), m.env, browserConfig{name: browser, path: path})
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterURL, nodeIdentifier, dev.DeviceID(), authToken))
		dev.AddActionHandler(remote.NewActionHandler(masterURL, nodeIdentifier, dev.DeviceID(), authToken))
		m.Add(dev)
		if err := m.deviceStorage.Update(m.Name(), dev); err != nil {
			return fmt.Errorf("store device %s: %w", dev.DeviceID(), err)
		}
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

// RefreshDevices does nothing, the browsers are registered once by Init.
func (m *Handler) RefreshDevices(bool) error { return nil }

func (m *Handler) RegisterDevice(device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device: %w", device.ErrUnsupported)
}
