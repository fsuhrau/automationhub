package iosdevice

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fsuhrau/automationhub/device"
)

var (
	DEVICE_LIST_TIMEOUT_SECS = 2
)

type IOSDevice struct {
	BuildVersion     string `json:"BuildVersion"`
	ModelSDK         string `json:"modelSDK"`
	DeviceIdentifier string `json:"DeviceIdentifier"`
	DeviceClass      string `json:"DeviceClass"`
	ProductType      string `json:"ProductType"`
	DeviceName       string `json:"DeviceName"`
	ProductVersion   string `json:"ProductVersion"`
	ModelArch        string `json:"modelArch"`
	HardwareModel    string `json:"HardwareModel"`
	ModelName        string `json:"modelName"`
}

type Detect struct {
	Event  string    `json:"Event"`
	Device IOSDevice `json:"Device"`
}

type Manager struct {
	devices map[string]*Device
}

func NewManager() *Manager {
	return &Manager{devices: make(map[string]*Device)}
}

func (m *Manager) Name() string {
	return "iosdevice"
}

func (m *Manager) Start() error {
	return nil
}

func (m *Manager) Stop() error {
	return nil
}

func (m *Manager) StartDevice(deviceID string) error {
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			found = true
			break
		}
	}

	if found {
		return nil
	}

	return device.DeviceNotFoundError
}

func (m *Manager) StopDevice(deviceID string) error {
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			found = true
			break
		}
	}

	if found {
		return nil
	}

	return device.DeviceNotFoundError
}

func (m *Manager) GetDevices() ([]device.Device, error) {
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Manager) RefreshDevices() error {
	cmd := device.NewCommand("/usr/local/bin/ios-deploy", "--detect", "--no-wifi", "--timeout", fmt.Sprintf("%d", DEVICE_LIST_TIMEOUT_SECS), "-j")
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	jsonString := fmt.Sprintf("[%s]", strings.Replace(string(output), "}{", "},{", -1))
	var resp []Detect
	if err := json.Unmarshal([]byte(jsonString), &resp); err != nil {
		return err
	}
	for _, device := range resp {
		if len(device.Device.BuildVersion) == 0 {
			continue
		}
		if _, ok := m.devices[device.Device.DeviceIdentifier]; ok {
			m.devices[device.Device.DeviceIdentifier].deviceName = device.Device.DeviceName
			m.devices[device.Device.DeviceIdentifier].deviceID = device.Device.DeviceIdentifier
			m.devices[device.Device.DeviceIdentifier].deviceOSName = device.Device.ModelSDK
			m.devices[device.Device.DeviceIdentifier].deviceOSVersion = device.Device.ProductVersion
			m.devices[device.Device.DeviceIdentifier].SetDeviceState("Booted")

		} else {
			m.devices[device.Device.DeviceIdentifier] = &Device{
				deviceName:      device.Device.DeviceName,
				deviceID:        device.Device.DeviceIdentifier,
				deviceOSName:    device.Device.ModelSDK,
				deviceOSVersion: device.Device.ProductVersion,
			}
			m.devices[device.Device.DeviceIdentifier].SetDeviceState("Booted")
		}
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
