package iossim

import (
	"encoding/json"
	"net"
	"regexp"

	"github.com/fsuhrau/automationhub/devices"
)

type SimDevice struct {
	State       string `json:"state"`
	IsAvailable bool   `json:"isAvailable"`
	Name        string `json:"name"`
	UDID        string `json:"udid"`
}

type SimulatoDescriptions struct {
	Devices map[string][]SimDevice `json:"devices"`
}

type Manager struct {
	devices map[string]*Device
	hostIP  net.IP
}

func NewManager(ip net.IP) *Manager {
	return &Manager{devices: make(map[string]*Device), hostIP: ip}
}

func (m *Manager) Name() string {
	return "iossim"
}

func (m *Manager) Start() error {
	cmd := devices.NewCommand("open", "-a", "/Applications/Xcode.app/Contents/Developer/Applications/Simulator.app/Contents/MacOS/Simulator")
	return cmd.Run()
}

func (m *Manager) Stop() error {
	cmd := devices.NewCommand("killall", "Simulator")
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
		cmd := devices.NewCommand("xcrun", "simctl", "boot", deviceID)
		return cmd.Run()
	}

	return devices.DeviceNotFoundError
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
		cmd := devices.NewCommand("xcrun", "simctl", "shutdown", deviceID)
		return cmd.Run()
	}

	return devices.DeviceNotFoundError
}

func (m *Manager) GetDevices() ([]devices.Device, error) {
	devices := make([]devices.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Manager) RefreshDevices() error {
	cmd := devices.NewCommand("xcrun", "simctl", "list", "devices", "--json")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	var resp SimulatoDescriptions
	if err := json.Unmarshal(output, &resp); err != nil {
		return err
	}
	regex := regexp.MustCompile(`com.apple.CoreSimulator.SimRuntime.([a-z]+OS)-([0-9]+-[0-9]+)`)

	for runtime, devices := range resp.Devices {
		subs := regex.FindAllStringSubmatch(runtime, -1)
		deviceOSName := subs[0][1]
		osVersion := subs[0][2]

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
				m.devices[device.UDID].SetDeviceState(device.State)

			} else {
				m.devices[device.UDID] = &Device{
					deviceName:      device.Name,
					deviceID:        device.UDID,
					deviceOSName:    deviceOSName,
					deviceOSVersion: osVersion,
					deviceIP:        m.hostIP,
				}
				m.devices[device.UDID].SetDeviceState(device.State)
			}
		}
	}

	return nil
}

func (m *Manager) isSimulationRunning(deviceID string) bool {
	if d, ok := m.devices[deviceID]; ok {
		return d.DeviceState() == devices.Booted
	}
	return false
}

func (m *Manager) HasDevice(dev devices.Device) bool {
	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}
