package iosdevice

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"time"

	"github.com/danielpaulus/go-ios/ios"
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
	devices       map[string]*Device
	deviceStorage storage.Device
}

func NewHandler(ds storage.Device) *Handler {
	return &Handler{devices: make(map[string]*Device), deviceStorage: ds}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string) error {
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}

	for i := range devs {
		deviceId := devs[i].DeviceIdentifier
		dev := &Device{}
		dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.deviceID))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.deviceID))
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
	var dev *Device
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			dev = v
			break
		}
	}

	if dev != nil {
		return dev.StartXCUITestRunner()
	}

	return device.DeviceNotFoundError
}

func (m *Handler) StopDevice(deviceID string) error {
	var dev *Device
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			dev = v
			break
		}
	}

	if dev != nil {
		return dev.StopXCUITestRunner()
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

func (m *Handler) RefreshDevices(force bool) error {
	lastUpdate := time.Now().UTC()
	_ = lastUpdate
	deviceList, err := ios.ListDevices()
	if err != nil {
		return err
	}
	for _, i := range deviceList.DeviceList {
		identifier := i.Properties.SerialNumber
		allValues, err := ios.GetValues(i)
		if err != nil {
			return err
		}
		if _, ok := m.devices[identifier]; ok {
			m.devices[identifier].deviceModel = allValues.Value.ProductType
			m.devices[identifier].deviceName = allValues.Value.DeviceName
			m.devices[identifier].deviceID = identifier
			m.devices[identifier].deviceOSName = allValues.Value.ProductName
			m.devices[identifier].deviceOSVersion = allValues.Value.ProductVersion
			m.devices[identifier].lastUpdateAt = lastUpdate
			m.devices[identifier].SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), m.devices[identifier])
		} else {
			m.devices[identifier] = &Device{
				deviceModel:     allValues.Value.ProductType,
				deviceName:      allValues.Value.DeviceName,
				deviceID:        identifier,
				deviceOSName:    allValues.Value.ProductName,
				deviceOSVersion: allValues.Value.ProductVersion,
				lastUpdateAt:    lastUpdate,
			}
			dev := models.Device{
				DeviceIdentifier: identifier,
				DeviceType:       models.DeviceTypePhone,
				Name:             allValues.Value.DeviceName,
				Manager:          Manager,
				PlatformType:     models.PlatformTypeiOS,
				OS:               allValues.Value.ProductName,
				OSVersion:        m.devices[identifier].deviceOSVersion,
				ConnectionParameter: &models.ConnectionParameter{
					ConnectionType: models.ConnectionTypeUSB,
				},
			}
			m.deviceStorage.NewDevice(m.Name(), dev)
			m.devices[identifier].SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), m.devices[identifier])
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
