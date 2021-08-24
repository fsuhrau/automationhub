package macos

import (
	"github.com/fsuhrau/automationhub/config"
	"net"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

type Handler struct {
	devices      map[string]*Device
	hostIP       net.IP
	deviceConfig config.Interface
}

func NewHandler(deviceConfig config.Interface, ip net.IP) *Handler {
	return &Handler{devices: make(map[string]*Device), hostIP: ip, deviceConfig: deviceConfig}
}

func (m *Handler) Name() string {
	return "macos"
}

func (m *Handler) Init() error {
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

func (m *Handler) RefreshDevices() error {
	if len(m.devices) == 0 {
		m.devices["54decb62-3993-4031-9c6a-18ce048cc63c"] = &Device{
			deviceName:      "MacOS",
			deviceID:        "54decb62-3993-4031-9c6a-18ce048cc63c",
			deviceOSName:    "MacOSX",
			deviceOSVersion: "10-14",
			deviceIP:        m.hostIP,
			lastUpdateAt:    time.Now().UTC(),
		}
		m.devices["54decb62-3993-4031-9c6a-18ce048cc63c"].SetDeviceState("Booted")
	}
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
