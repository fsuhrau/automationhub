package iossim

import (
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/config"

	"github.com/fsuhrau/automationhub/device"
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

type Manager struct {
	devices      map[string]*Device
	hostIP       net.IP
	deviceConfig config.Interface
}

func NewManager(deviceConfig config.Interface, ip net.IP) *Manager {
	return &Manager{devices: make(map[string]*Device), hostIP: ip, deviceConfig: deviceConfig}
}

func (m *Manager) Name() string {
	return "ios_sim"
}

func (m *Manager) Init() error {
	return nil
}

func (m *Manager) Start() error {
	cmd := device.NewCommand("open", "-a", "/Applications/Xcode.app/Contents/Developer/Applications/Simulator.app/Contents/MacOS/Simulator")
	return cmd.Run()
}

func (m *Manager) Stop() error {
	cmd := device.NewCommand("killall", "Simulator")
	return cmd.Run()
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
		cmd := device.NewCommand("xcrun", "simctl", "boot", deviceID)
		return cmd.Run()
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
		cmd := device.NewCommand("xcrun", "simctl", "shutdown", deviceID)
		return cmd.Run()
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
	lastUpdate := time.Now().UTC()
	cmd := device.NewCommand("xcrun", "simctl", "list", "devices", "--json")
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
				m.devices[device.UDID].deviceIP = m.hostIP
				m.devices[device.UDID].lastUpdateAt = lastUpdate
				m.devices[device.UDID].SetDeviceState(device.State)

			} else {
				var cfg *config.Device
				if m.deviceConfig != nil {
					cfg = m.deviceConfig.GetDeviceConfig(device.UDID)
				}

				m.devices[device.UDID] = &Device{
					deviceName:      device.Name,
					deviceID:        device.UDID,
					deviceOSName:    deviceOSName,
					deviceOSVersion: osVersion,
					deviceIP:        m.hostIP,
					cfg:             cfg,
					lastUpdateAt:    lastUpdate,
				}
				m.devices[device.UDID].SetDeviceState(device.State)
			}
		}
	}

	for i := range m.devices {
		if m.devices[i].lastUpdateAt != lastUpdate {
			if m.devices[i].cfg != nil && m.devices[i].cfg.Connection.Type == "remote" {
				m.devices[i].SetDeviceState("RemoteDisconnected")
			} else {
				m.devices[i].SetDeviceState("Unknown")
			}
		}
	}

	return nil
}

func (m *Manager) isSimulationRunning(deviceID string) bool {
	if d, ok := m.devices[deviceID]; ok {
		return d.DeviceState() == device.Booted
	}
	return false
}

func (m *Manager) HasDevice(dev device.Device) bool {
	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}
