package iossim

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tools/exec"
	"regexp"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	Manager = "ios_sim"
)

var OSVersionLookupRegex = regexp.MustCompile(`com.apple.CoreSimulator.SimRuntime.([a-z]+OS)-([0-9]+-[0-9]+)`)

type SimDevice struct {
	State       string `json:"state"`
	IsAvailable bool   `json:"isAvailable"`
	Name        string `json:"name"`
	UDID        string `json:"udid"`
}

type SimulatorDescriptions struct {
	Devices map[string][]SimDevice `json:"devices"`
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
	cmd := exec.NewCommand("open", "-a", "/Applications/Xcode.app/Contents/Developer/Applications/Simulator.app/Contents/MacOS/Simulator")
	return cmd.Run()
}

func (m *Handler) Stop() error {
	cmd := exec.NewCommand("killall", "Simulator")
	return cmd.Run()
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
		cmd := exec.NewCommand("xcrun", "simctl", "boot", deviceID)
		return cmd.Run()
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
		cmd := exec.NewCommand("xcrun", "simctl", "shutdown", deviceID)
		return cmd.Run()
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
	cmd := exec.NewCommand("xcrun", "simctl", "list", "devices", "--json")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	var resp SimulatorDescriptions
	if err := json.Unmarshal(output, &resp); err != nil {
		return err
	}
	for runtime, devices := range resp.Devices {
		subs := OSVersionLookupRegex.FindAllStringSubmatch(runtime, -1)
		deviceOSName := subs[0][1]
		osVersion := strings.ReplaceAll(subs[0][2], "-", ".")

		if deviceOSName == "iOS" {
			deviceOSName = "iphonesimulator"
		}

		for _, device := range devices {
			if _, ok := m.devices[device.UDID]; ok {
				m.devices[device.UDID].deviceName = device.Name
				m.devices[device.UDID].deviceID = device.UDID
				m.devices[device.UDID].deviceOSName = deviceOSName
				m.devices[device.UDID].deviceOSVersion = osVersion
				m.devices[device.UDID].lastUpdateAt = lastUpdate
				m.devices[device.UDID].SetDeviceState(device.State)
				m.deviceStorage.Update(m.Name(), m.devices[device.UDID])
			} else {
				m.devices[device.UDID] = &Device{
					deviceName:      device.Name,
					deviceID:        device.UDID,
					deviceOSName:    deviceOSName,
					deviceOSVersion: osVersion,
					lastUpdateAt:    lastUpdate,
				}
				m.devices[device.UDID].SetDeviceState(device.State)
				dev := models.Device{
					DeviceIdentifier: device.UDID,
					DeviceType:       models.DeviceTypePhone,
					Name:             device.Name,
					Manager:          Manager,
					PlatformType:     models.PlatformTypeiOSSimulator,
					OS:               deviceOSName,
					OSVersion:        osVersion,
					ConnectionParameter: &models.ConnectionParameter{
						ConnectionType: models.ConnectionTypeUSB,
					},
				}
				m.deviceStorage.NewDevice(m.Name(), dev)
				m.deviceStorage.Update(m.Name(), m.devices[device.UDID])
			}
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

func (m *Handler) isSimulationRunning(deviceID string) bool {
	if d, ok := m.devices[deviceID]; ok {
		return d.DeviceState() == device.StateBooted
	}
	return false
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
