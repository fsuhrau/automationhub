package androiddevice

import (
	"bufio"
	"bytes"
	"github.com/fsuhrau/automationhub/config"
	"os"
	"regexp"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var DeviceListRegex = regexp.MustCompile(`([a-zA-Z0-9.:]+)\s+device\s(usb:([a-zA-Z0-9]+)\s|)product:([a-zA-Z0-9_]+)\smodel:([a-zA-Z0-9_]+)\s+device:([a-zA-Z0-9]+)\s+transport_id:([0-9]+)`)

type Manager struct {
	devices map[string]*Device
	deviceConfig config.Interface
}

func NewManager(deviceConfig config.Interface) *Manager {
	return &Manager{devices: make(map[string]*Device), deviceConfig: deviceConfig}
}

func (m *Manager) Name() string {
	return "androiddevice"
}

func (m *Manager) Init() error {
	devices := m.deviceConfig.GetDevicesForManager(m.Name())
	for i := range devices {
		// connect all remote devices
		if devices[i].Connection.Type == "remote" {
			cmd := device.NewCommand("adb", "connect", devices[i].Connection.IP)
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	if err := m.RefreshDevices(); err != nil {
		return err
	}

	for i := range devices {
		// connect all remote devices
		if devices[i].Connection.Type == "remote" {
			cmd := device.NewCommand("adb", "disconnect", devices[i].Connection.IP)
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) Start() error {
	cmd := device.NewCommand("adb", "start-server")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) Stop() error {
	cmd := device.NewCommand("adb", "kill-server")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) StartDevice(deviceID string) error {
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			if v.deviceState == device.RemoteDisconnected {
				cmd := device.NewCommand("adb", "connect", v.cfg.Connection.IP)
				cmd.Stderr = os.Stderr
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

func (m *Manager) StopDevice(deviceID string) error {
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			found = true
			cmd := device.NewCommand("adb", "disconnect", v.cfg.Connection.IP)
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
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
	lastUpdate := time.Now().UTC()
	cmd := device.NewCommand("adb", "devices", "-l")
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
			m.devices[deviceID].deviceName = name
			m.devices[deviceID].deviceID = deviceID
			m.devices[deviceID].SetDeviceState("Booted")
			m.devices[deviceID].product = product
			m.devices[deviceID].deviceUSB = deviceUSB
			m.devices[deviceID].deviceModel = model
			m.devices[deviceID].transportID = transportID
			m.devices[deviceID].lastUpdateAt = lastUpdate

		} else {
			var cfg *config.Device
			if m.deviceConfig != nil {
				cfg = m.deviceConfig.GetDeviceConfig(deviceID)
			}
			m.devices[deviceID] = &Device{
				deviceName:  name,
				deviceID:    deviceID,
				product:     product,
				deviceUSB:   deviceUSB,
				deviceModel: model,
				transportID: transportID,
				deviceOSName: "android",
				cfg: cfg,
				lastUpdateAt: lastUpdate,
			}
			m.devices[deviceID].UpdateDeviceInfos()
			m.devices[deviceID].SetDeviceState("Booted")
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

func (m *Manager) HasDevice(dev device.Device) bool {
	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}

// public void createUsbTunnel(String deviceId, int localPort, int remotePort) {
// 	BlockingProcess process = new BlockingProcess((n, line) -> LOGGER.debug(line),
// 		"adb", "-s", deviceId, "forward", "tcp:" + localPort, "tcp:" + remotePort);

// 	if (process.hasExited() && process.getExitValue() != 0) {
// 		LOGGER.error("Cannot open USB tunnel for {} with src: {} dst: {}, see the log for error messages", deviceId, localPort, remotePort);
// 	}
// }

// public void stopTunnel(String deviceId, int localPort) {
// 	new BlockingProcess((n, line) -> LOGGER.debug(line),
// 		"adb", "-s", deviceId, "forward", "--remove", String.valueOf(localPort));
// }

// @Override
// public InetSocketAddress getInetSocketAddress(String deviceUuid) throws IOException {
// 	int localPort = NetworkUtility.findUnusedLocalPort();
// 	return new InetSocketAddress(InetAddress.getLoopbackAddress(), localPort);
// }
