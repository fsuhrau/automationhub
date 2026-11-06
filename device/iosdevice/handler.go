package iosdevice

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

const (
	Manager = "ios_device"
)

var (
	DeviceListTimeoutSecs = 2
)

type IOSDevice struct {
	BuildVersion     string `json:"BuildVersion"`
	ModelSDK         string `json:"ModelSDK"`
	DeviceIdentifier string `json:"DeviceIdentifier"`
	DeviceClass      string `json:"DeviceClass"`
	ProductType      string `json:"ProductType"`
	DeviceName       string `json:"DeviceName"`
	ProductVersion   string `json:"ProductVersion"`
	ModelArch        string `json:"ModelArch"`
	HardwareModel    string `json:"HardwareModel"`
	ModelName        string `json:"ModelName"`
}

type Detect struct {
	Event  string    `json:"Event"`
	Device IOSDevice `json:"Device"`
}

// fallbackPollInterval: while the usbmuxd event subscription is up, the
// device list is still polled at this interval as a safety net.
const fallbackPollInterval = time.Minute

// settleDelay: a freshly attached iPhone often can't be read right away, so
// it is read again after this delay.
const settleDelay = 3 * time.Second

type Handler struct {
	mu            sync.Mutex // guards devices and lastPoll
	refreshMu     sync.Mutex // serializes refreshes
	devices       map[string]*Device
	deviceStorage storage.Device

	masterURL      string
	nodeIdentifier string
	authToken      *string

	notifier  *node.DevicesNotifier
	tracking  atomic.Bool
	lastPoll  time.Time
	stopWatch context.CancelFunc
}

func NewHandler(ds storage.Device) *Handler {
	return &Handler{devices: make(map[string]*Device), deviceStorage: ds}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	m.masterURL = masterUrl
	m.nodeIdentifier = nodeIdentifier
	m.authToken = authToken
	m.notifier = node.NewDevicesNotifier(masterUrl, nodeIdentifier, authToken)
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}

	for i := range devs {
		deviceId := devs[i].DeviceIdentifier
		dev := &Device{deviceID: deviceId}
		dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.deviceID, authToken))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.deviceID, authToken))
		m.devices[deviceId] = dev
	}
	return nil
}

// Start subscribes to usbmuxd attach/detach events, so iPhones show up and
// disappear right away instead of with the next poll.
func (m *Handler) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.stopWatch = cancel
	m.mu.Unlock()
	go watchUsbmux(ctx, openUsbmuxListener, func() {
		m.refreshAndNotify()
		// read again once a freshly attached device is ready
		time.AfterFunc(settleDelay, func() {
			if ctx.Err() == nil {
				m.refreshAndNotify()
			}
		})
	}, m.tracking.Store)
	return nil
}

func (m *Handler) refreshAndNotify() {
	if err := m.RefreshDevices(true); err != nil {
		logrus.Debugf("ios device refresh: %v", err)
	}
	m.notifier.Notify()
}

func (m *Handler) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopWatch != nil {
		m.stopWatch()
		m.stopWatch = nil
	}
	return nil
}

func (m *Handler) StartDevice(deviceID string) error {
	m.mu.Lock()
	var dev *Device
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			dev = v
			break
		}
	}

	m.mu.Unlock()
	if dev != nil {
		return dev.StartXCUITestRunner()
	}

	return device.DeviceNotFoundError
}

func (m *Handler) StopDevice(deviceID string) error {
	m.mu.Lock()
	var dev *Device
	for _, v := range m.devices {
		if v.deviceID == deviceID {
			dev = v
			break
		}
	}

	m.mu.Unlock()
	if dev != nil {
		return dev.StopXCUITestRunner()
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

// RefreshDevices reads the attached devices from usbmuxd and lockdown. While
// the usbmuxd event subscription is up, unforced calls only run as a safety
// net every fallbackPollInterval.
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

	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	deviceList, err := ios.ListDevices()
	if err != nil {
		return err
	}

	// read the device values without holding the device lock (lockdown calls)
	type attached struct {
		serial string
		values *ios.GetAllValuesResponse
		err    error
	}
	var found []attached
	for _, entry := range deviceList.DeviceList {
		values, err := ios.GetValues(entry)
		if err != nil {
			found = append(found, attached{serial: entry.Properties.SerialNumber, err: err})
			continue
		}
		found = append(found, attached{serial: entry.Properties.SerialNumber, values: &values})
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	lastUpdate := time.Now().UTC()
	for _, a := range found {
		dev, known := m.devices[a.serial]
		if a.err != nil {
			// attached but lockdown refuses: not trusted yet, or locked with a passcode
			logrus.Debugf("ios device %s not readable: %v", a.serial, a.err)
			if !known {
				dev = &Device{deviceID: a.serial, deviceName: a.serial, deviceOSName: "iOS"}
				m.devices[a.serial] = dev
				m.registerNewDevice(dev, nil)
			}
			dev.deviceID = a.serial
			dev.lastUpdateAt = lastUpdate
			dev.SetDeviceState("StateUnauthorized")
			m.deviceStorage.Update(m.Name(), dev)
			continue
		}
		if known {
			dev.deviceID = a.serial
			dev.lastUpdateAt = lastUpdate
			dev.UpdateDeviceInfos(*a.values)
			dev.SetDeviceState("StateBooted")
			m.deviceStorage.Update(m.Name(), dev)
			continue
		}
		dev = &Device{deviceID: a.serial, lastUpdateAt: lastUpdate}
		dev.UpdateDeviceInfos(*a.values)
		dev.SetDeviceState("StateBooted")
		m.devices[a.serial] = dev
		m.registerNewDevice(dev, a.values)
	}

	// known devices that aren't attached (unplugged, network devices away)
	for _, dev := range m.devices {
		if dev.lastUpdateAt != lastUpdate {
			dev.SetDeviceState("StateRemoteDisconnected")
			m.deviceStorage.Update(m.Name(), dev)
		}
	}
	return nil
}

// registerNewDevice stores a newly seen device and wires its remote logging.
func (m *Handler) registerNewDevice(d *Device, values *ios.GetAllValuesResponse) {
	dev := models.Device{
		DeviceIdentifier: d.deviceID,
		DeviceType:       models.DeviceTypePhone,
		Name:             d.deviceName,
		Manager:          Manager,
		PlatformType:     models.PlatformTypeiOS,
		OS:               d.deviceOSName,
		OSVersion:        d.deviceOSVersion,
		ConnectionParameter: &models.ConnectionParameter{
			ConnectionType: models.ConnectionTypeUSB,
		},
	}
	if values != nil {
		dev.Name = values.Value.DeviceName
		dev.OS = values.Value.ProductName
	}
	d.SetLogWriter(generic.NewRemoteLogWriter(m.masterURL, m.nodeIdentifier, d.deviceID, m.authToken))
	d.AddActionHandler(node.NewRemoteActionHandler(m.masterURL, m.nodeIdentifier, d.deviceID, m.authToken))
	m.deviceStorage.NewDevice(m.Name(), dev)
	m.deviceStorage.Update(m.Name(), d)
	d.SetConfig(&dev)
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
