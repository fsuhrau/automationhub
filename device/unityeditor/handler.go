package unityeditor

import (
	"github.com/fsuhrau/automationhub/storage"
	"net"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	Manager = "unity_editor"
)

type Handler struct {
	devices      map[string]*Device
	hostIP       net.IP
	deviceStorage storage.Device
}

func NewHandler(ds storage.Device, ip net.IP) *Handler {
	return &Handler{devices: make(map[string]*Device), hostIP: ip, deviceStorage: ds}
}

func (m *Handler) Name() string {
	return Manager
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
	now := time.Now().UTC()

	for i := range m.devices {
		if now.Sub(m.devices[i].lastUpdateAt) > 1*time.Minute {
			if m.devices[i].deviceState != device.StateUnknown {
				m.devices[i].deviceState = device.StateUnknown
				m.devices[i].updated = true
			}
		}
		if m.devices[i].updated {
			m.deviceStorage.Update(m.Name(), m.devices[i])
			m.devices[i].updated = false
		}
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

func (m *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	lastUpdate := time.Now().UTC()

	if _, ok := m.devices[data.DeviceID]; ok {
		m.devices[data.DeviceID].deviceOSName = data.DeviceOS
		m.devices[data.DeviceID].deviceOSVersion = data.DeviceOSVersion
		m.devices[data.DeviceID].deviceName = data.Name
		m.devices[data.DeviceID].deviceIP = data.DeviceIP
		m.devices[data.DeviceID].lastUpdateAt = lastUpdate
		m.devices[data.DeviceID].conn = data.Conn
		m.devices[data.DeviceID].updated = true
	} else {
		m.devices[data.DeviceID] = &Device{
			deviceName:      data.Name,
			deviceID:        data.DeviceID,
			deviceOSName:    data.DeviceOS,
			deviceOSVersion: data.DeviceOSVersion,
			deviceIP:        data.DeviceIP,
			lastUpdateAt:    lastUpdate,
			conn:            data.Conn,
			updated:         true,
		}
	}
	go m.devices[data.DeviceID].HandleSocketFunction()
	return m.devices[data.DeviceID], nil
}
