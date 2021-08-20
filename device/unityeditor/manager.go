package unityeditor

import (
	"github.com/fsuhrau/automationhub/config"
	"net"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

type Manager struct {
	devices      map[string]*Device
	hostIP       net.IP
	deviceConfig config.Interface
}

func NewManager(deviceConfig config.Interface, ip net.IP) *Manager {
	return &Manager{devices: make(map[string]*Device), hostIP: ip, deviceConfig: deviceConfig}
}

func (m *Manager) Name() string {
	return "unity_editor"
}

func (m *Manager) Init() error {
	return nil
}

func (m *Manager) Start() error {
	return nil
}

func (m *Manager) Stop() error {
	return nil
}

func (m *Manager) StartDevice(deviceID string) error {
	return nil
}

func (m *Manager) StopDevice(deviceID string) error {
	return nil
}

func (m *Manager) GetDevices() ([]device.Device, error) {
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Manager) RefreshDevices() error {
	if len(m.devices) == 0 {
		m.devices["dd7ace2f-07b8-4696-92b7-a856dc0c04b5"] = &Device{
			deviceName:      "UnityEditor",
			deviceID:        "dd7ace2f-07b8-4696-92b7-a856dc0c04b5",
			deviceOSName:    "UnityEditor",
			deviceOSVersion: "version",
			deviceIP:        m.hostIP,
			lastUpdateAt:    time.Now().UTC(),
		}
		m.devices["dd7ace2f-07b8-4696-92b7-a856dc0c04b5"].SetDeviceState("Booted")
	}
	return nil
}

func (m *Manager) HasDevice(dev device.Device) bool {
	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}
