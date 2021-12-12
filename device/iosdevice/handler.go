package iosdevice

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tools/exec"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	Manager = "ios_device"
)

var (
	DeviceListTimeoutSecs = 2
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

type Handler struct {
	devices      map[string]*Device
	deviceStorage storage.Device
}

func NewHandler(ds storage.Device) *Handler {
	return &Handler{devices: make(map[string]*Device), deviceStorage: ds}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init() error {
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}

	for i := range devs {
		deviceId := devs[i].DeviceIdentifier
		dev := &Device{}
		dev.SetConfig(&devs[i])
		m.devices[deviceId] = dev
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

func (m *Handler) StopDevice(deviceID string) error {
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

func (m *Handler) GetDevices() ([]device.Device, error) {
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Handler) RefreshDevices() error {
	lastUpdate := time.Now().UTC()
	cmd := exec.NewCommand(IosDeployBin, "--detect", "--no-wifi", "--timeout", fmt.Sprintf("%d", DeviceListTimeoutSecs), "-j")
	// cmd.Stderr = os.Stderr
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
			m.devices[device.Device.DeviceIdentifier].lastUpdateAt = lastUpdate
			m.devices[device.Device.DeviceIdentifier].SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), m.devices[device.Device.DeviceIdentifier])
		} else {
			m.devices[device.Device.DeviceIdentifier] = &Device{
				deviceName:      device.Device.DeviceName,
				deviceID:        device.Device.DeviceIdentifier,
				deviceOSName:    device.Device.ModelSDK,
				deviceOSVersion: device.Device.ProductVersion,
				lastUpdateAt:    lastUpdate,
			}
			m.devices[device.Device.DeviceIdentifier].SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), m.devices[device.Device.DeviceIdentifier])
		}
	}

	for i := range m.devices {
		if m.devices[i].lastUpdateAt != lastUpdate {
			if m.devices[i].GetConfig() != nil && m.devices[i].GetConfig().ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
				m.devices[i].SetDeviceState("StateRemoteDisconnected")
			} else {
				m.devices[i].SetDeviceState("StateUnknown")
			}
			m.deviceStorage.Update(m.Name(), m.devices[i])
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
	return nil, fmt.Errorf("register device not implemented")
}
