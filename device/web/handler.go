package web

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
)

const (
	Manager = "web"
)

type Handler struct {
	devices        map[string]*Device
	deviceStorage  storage.Device
	init           bool
	masterURL      string
	nodeIdentifier string
	cfg            config.Manager
}

func NewHandler(cfg config.Manager, ds storage.Device) *Handler {
	return &Handler{cfg: cfg, devices: make(map[string]*Device), deviceStorage: ds}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {

	m.init = true

	m.masterURL = masterUrl
	m.nodeIdentifier = nodeIdentifier

	defer func() {
		m.init = false
	}()

	for k, v := range m.cfg.Browser {
		dev := &Device{
			browser:     k,
			browserPath: v,
		}

		dev.UpdateDeviceInfos()
		//dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.DeviceID(), authToken))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.DeviceID(), authToken))
		dev.SetDeviceState("StateBooted")
		m.devices[dev.DeviceID()] = dev
		m.deviceStorage.Update(m.Name(), dev)
	}

	if err := m.RefreshDevices(true); err != nil {
		return err
	}
	return nil
}

func (m *Handler) Start() error {
	return nil
}

func (m *Handler) Stop() error {
	return nil
}

func (m *Handler) StartDevice(deviceID string) error {
	return nil
}

func (m *Handler) StopDevice(deviceID string) error {
	return nil
}

func (m *Handler) GetDevices() ([]device.Device, error) {
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Handler) RefreshDevices(force bool) error {
	/*
		if len(m.devices) == 0 {
			m.devices["54decb62-3993-4031-9c6a-18ce048cc63c"] = &Device{
				deviceName:      "MacOS",
				deviceID:        "54decb62-3993-4031-9c6a-18ce048cc63c",
				deviceOSName:    "MacOSX",
				deviceOSVersion: "10-14",
				lastUpdateAt:    time.Now().UTC(),
			}
			m.devices["54decb62-3993-4031-9c6a-18ce048cc63c"].SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), m.devices["54decb62-3993-4031-9c6a-18ce048cc63c"])
		}
	*/
	return nil
}

func (m *Handler) HasDevice(dev device.Device) bool {
	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}

func (m *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device not implemented")
}
