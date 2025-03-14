package androiddevice

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tools/exec"
	"regexp"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var DeviceListRegex = regexp.MustCompile(`([a-zA-Z0-9.:\-]+)\s+device\s(usb:([a-zA-Z0-9]+)\s|)product:([a-zA-Z0-9_]+)\smodel:([a-zA-Z0-9_]+)\s+device:([a-zA-Z0-9]+)\s+transport_id:([0-9]+)`)

const (
	Manager = "android_device"
)

type Handler struct {
	devices        map[string]*Device
	deviceStorage  storage.Device
	init           bool
	masterURL      string
	nodeIdentifier string
}

func NewHandler(ds storage.Device) *Handler {
	return &Handler{
		devices:       make(map[string]*Device),
		deviceStorage: ds,
	}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string) error {
	m.init = true
	m.masterURL = masterUrl
	m.nodeIdentifier = nodeIdentifier
	defer func() {
		m.init = false
	}()
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for i := range devs {
		deviceId := devs[i].DeviceIdentifier
		dev := &Device{
			deviceOSName:  "android",
			installedApps: make(map[string]string),
		}
		dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.deviceID))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.deviceID))
		m.devices[deviceId] = dev

		// connect all remote devs
		if devs[i].ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
			connectionString := GetConnectionString(devs[i].ConnectionParameter)
			cmd := exec.NewCommand("adb", "connect", connectionString)
			// cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	if err := m.RefreshDevices(true); err != nil {
		return err
	}

	for i := range devs {
		// connect all remote devs
		if devs[i].ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
			connectionString := GetConnectionString(devs[i].ConnectionParameter)
			cmd := exec.NewCommand("adb", "disconnect", connectionString)
			// cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Handler) Start() error {
	cmd := exec.NewCommand("adb", "start-server")
	// cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Handler) Stop() error {
	cmd := exec.NewCommand("adb", "kill-server")
	// cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Handler) StartDevice(deviceID string) error {
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			if v.deviceState == device.StateRemoteDisconnected {
				dev, err := m.deviceStorage.GetDevice(Manager, deviceID)
				if err != nil {
					return device.DeviceNotFoundError
				}
				cmd := exec.NewCommand("adb", "connect", GetConnectionString(dev.ConnectionParameter))
				// cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
			}
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
			if v.GetConfig() != nil && v.GetConfig().ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
				cmd := exec.NewCommand("adb", "disconnect", GetConnectionString(v.GetConfig().ConnectionParameter))
				// cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
			}
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

func (m *Handler) RefreshDevices(force bool) error {
	lastUpdate := time.Now().UTC()
	cmd := exec.NewCommand("adb", "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := DeviceListRegex.FindAllStringSubmatch(line, -1)
		if len(matches) < 1 {
			continue
		}
		if len(matches[0]) < 7 {
			continue
		}

		deviceID := matches[0][1]
		deviceUSB := matches[0][2]
		product := matches[0][3]
		model := matches[0][4]
		name := matches[0][5]
		transportID := matches[0][6]

		if _, ok := m.devices[deviceID]; ok {
			dev := m.devices[deviceID]
			dev.deviceName = name
			dev.deviceID = deviceID
			dev.product = product
			dev.deviceUSB = deviceUSB
			dev.deviceModel = model
			dev.transportID = transportID
			dev.lastUpdateAt = lastUpdate
			dev.SetDeviceState("StateBooted")
			if m.init {
				m.devices[deviceID].UpdateDeviceInfos()
			}
			m.deviceStorage.Update(m.Name(), dev)
		} else {
			m.devices[deviceID] = &Device{
				deviceName:    name,
				deviceID:      deviceID,
				product:       product,
				deviceUSB:     deviceUSB,
				deviceModel:   model,
				transportID:   transportID,
				deviceOSName:  "android",
				lastUpdateAt:  lastUpdate,
				installedApps: make(map[string]string),
			}
			m.devices[deviceID].UpdateDeviceInfos()
			m.devices[deviceID].SetDeviceState("StateBooted")
			dev := models.Device{
				DeviceIdentifier: deviceID,
				DeviceType:       models.DeviceTypePhone,
				Name:             name,
				Manager:          Manager,
				HardwareModel:    model,
				OS:               "android",
				PlatformType:     models.PlatformTypeAndroid,
				OSVersion:        m.devices[deviceID].deviceOSVersion,
				ConnectionParameter: &models.ConnectionParameter{
					ConnectionType: models.ConnectionTypeUSB,
				},
			}

			m.devices[deviceID].SetLogWriter(generic.NewRemoteLogWriter(m.masterURL, m.nodeIdentifier, deviceID))
			m.devices[deviceID].AddActionHandler(node.NewRemoteActionHandler(m.masterURL, m.nodeIdentifier, deviceID))

			m.deviceStorage.NewDevice(m.Name(), dev)
			m.deviceStorage.Update(m.Name(), m.devices[deviceID])
			m.devices[deviceID].Config = &dev
		}
	}

	for i := range m.devices {
		if m.devices[i].lastUpdateAt != lastUpdate {
			if m.devices[i].GetConfig() != nil && m.devices[i].GetConfig().ConnectionParameter.ConnectionType == models.ConnectionTypeRemote {
				m.devices[i].SetDeviceState("StateRemoteDisconnected")
				m.deviceStorage.Update(m.Name(), m.devices[i])
			} else {
				m.devices[i].SetDeviceState("StateUnknown")
				m.deviceStorage.Update(m.Name(), m.devices[i])
			}
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
