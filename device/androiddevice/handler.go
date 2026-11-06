package androiddevice

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tools/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	Manager = "android_device"
)

// fallbackPollInterval: while the adb event stream is connected, the device
// list is still polled at this interval as a safety net.
const fallbackPollInterval = time.Minute

type Handler struct {
	mu             sync.Mutex // guards devices and lastPoll
	devices        map[string]*Device
	deviceStorage  storage.Device
	init           bool
	masterURL      string
	nodeIdentifier string
	authToken      *string

	adbServerAddr string
	notifier      *node.DevicesNotifier
	tracking      atomic.Bool // the adb track-devices stream is connected
	lastPoll      time.Time
	stopWatch     context.CancelFunc
}

func NewHandler(ds storage.Device) *Handler {
	return &Handler{
		devices:       make(map[string]*Device),
		deviceStorage: ds,
		adbServerAddr: DefaultADBServerAddr,
	}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	m.authToken = authToken
	m.init = true
	m.masterURL = masterUrl
	m.nodeIdentifier = nodeIdentifier
	m.notifier = node.NewDevicesNotifier(masterUrl, nodeIdentifier, authToken)
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
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.deviceID, authToken))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.deviceID, authToken))
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

func startADBServer() error {
	return exec.NewCommand("adb", "start-server").Run()
}

// Start starts the adb server and subscribes to its device events, so
// devices that are plugged in, unplugged or authorized show up right away.
func (m *Handler) Start() error {
	err := startADBServer()

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.stopWatch = cancel
	m.mu.Unlock()
	go watchADBDevices(ctx, m.adbServerAddr, startADBServer, func(entries []adbDeviceEntry) {
		m.applyDeviceList(entries)
		m.notifier.Notify()
	}, m.tracking.Store)
	return err
}

func (m *Handler) Stop() error {
	m.mu.Lock()
	if m.stopWatch != nil {
		m.stopWatch()
		m.stopWatch = nil
	}
	m.mu.Unlock()
	cmd := exec.NewCommand("adb", "kill-server")
	// cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Handler) StartDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	found := false
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			if v.deviceState == device.StateRemoteDisconnected {
				dev, err := m.deviceStorage.GetDevice(Manager, deviceID)
				if err != nil {
					return device.DeviceNotFoundError
				}
				// only network devices can be reconnected; an unplugged USB
				// device can't be started from here
				if dev.ConnectionParameter == nil || dev.ConnectionParameter.ConnectionType != models.ConnectionTypeRemote {
					found = true
					break
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
	m.mu.Lock()
	defer m.mu.Unlock()
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
	m.mu.Lock()
	defer m.mu.Unlock()
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

// registerNewDevice stores a newly seen device and wires its remote logging.
func (m *Handler) registerNewDevice(d *Device) {
	dev := models.Device{
		DeviceIdentifier: d.deviceID,
		DeviceType:       models.DeviceTypePhone,
		Name:             d.deviceName,
		Manager:          Manager,
		OS:               "android",
		PlatformType:     models.PlatformTypeAndroid,
		OSVersion:        d.deviceOSVersion,
		ConnectionParameter: &models.ConnectionParameter{
			ConnectionType: models.ConnectionTypeUSB,
		},
	}

	d.SetLogWriter(generic.NewRemoteLogWriter(m.masterURL, m.nodeIdentifier, d.deviceID, m.authToken))
	d.AddActionHandler(node.NewRemoteActionHandler(m.masterURL, m.nodeIdentifier, d.deviceID, m.authToken))

	m.deviceStorage.NewDevice(m.Name(), dev)
	m.deviceStorage.Update(m.Name(), d)
	d.SetConfig(&dev)
}

// updateUnavailableDevice handles adb entries that are listed but not usable,
// e.g. "unauthorized" (USB debugging not accepted on the phone) or "offline".
func (m *Handler) updateUnavailableDevice(entry adbDeviceEntry, lastUpdate time.Time) {
	state := "StateRemoteDisconnected"
	if entry.State == adbStateUnauthorized {
		state = "StateUnauthorized"
	}
	if d, ok := m.devices[entry.Serial]; ok {
		d.deviceID = entry.Serial
		d.lastUpdateAt = lastUpdate
		d.SetDeviceState(state)
		m.deviceStorage.Update(m.Name(), d)
		return
	}
	if entry.State != adbStateUnauthorized {
		return
	}
	// a new phone waiting for authorization: list it, but its properties can
	// only be read once it is authorized (deviceParameter stays nil until then)
	d := &Device{
		deviceID:      entry.Serial,
		deviceName:    entry.Serial,
		deviceOSName:  "android",
		lastUpdateAt:  lastUpdate,
		installedApps: make(map[string]string),
	}
	d.SetDeviceState(state)
	m.devices[entry.Serial] = d
	m.registerNewDevice(d)
}

// RefreshDevices polls "adb devices -l". While the adb event stream is
// connected this only runs as a safety net every fallbackPollInterval.
func (m *Handler) RefreshDevices(force bool) error {
	m.mu.Lock()
	skip := !force && m.tracking.Load() && time.Since(m.lastPoll) < fallbackPollInterval
	if !skip {
		m.lastPoll = time.Now()
	}
	m.mu.Unlock()
	if skip {
		return nil
	}

	output, err := exec.NewCommand("adb", "devices", "-l").Output()
	if err != nil {
		return err
	}
	m.applyDeviceList(parseADBDeviceList(string(output)))
	return nil
}

// applyDeviceList updates all devices from a complete adb device list.
func (m *Handler) applyDeviceList(entries []adbDeviceEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lastUpdate := time.Now().UTC()
	devs, _ := m.deviceStorage.GetDevices(Manager)
	for _, entry := range entries {
		if entry.State != adbStateDevice {
			m.updateUnavailableDevice(entry, lastUpdate)
			continue
		}
		deviceID := entry.Serial
		if dev, ok := m.devices[deviceID]; ok {
			// devices loaded from storage start without id and config
			dev.deviceID = deviceID
			if dev.GetConfig() == nil {
				for _, d := range devs {
					if d.DeviceIdentifier == deviceID {
						dev.SetConfig(d)
					}
				}
			}
			dev.lastUpdateAt = lastUpdate
			// reads the device properties on first use (e.g. after authorization)
			dev.UpdateDeviceInfos(entry)
			dev.SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), dev)
		} else {
			m.devices[deviceID] = &Device{
				deviceID:      deviceID,
				deviceOSName:  "android",
				lastUpdateAt:  lastUpdate,
				installedApps: make(map[string]string),
			}
			m.devices[deviceID].UpdateDeviceInfos(entry)
			m.devices[deviceID].SetDeviceState("StateBooted")
			m.registerNewDevice(m.devices[deviceID])
		}
	}

	// known devices adb doesn't list: unplugged, emulator stopped, network
	// device not connected
	for i := range m.devices {
		if m.devices[i].lastUpdateAt != lastUpdate {
			m.devices[i].SetDeviceState("StateRemoteDisconnected")
			m.deviceStorage.Update(m.Name(), m.devices[i])
		}
	}
}

func (m *Handler) HasDevice(dev device.Device) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
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
