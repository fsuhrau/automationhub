package iosdevice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/modules/webdriver"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
)

const (
	Manager = "ios_device"
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

// Handler manages the iOS devices of this host. The embedded registry gives
// GetDevices and HasDevice.
type Handler struct {
	generic.Registry[*Device]
	deps          Deps
	wdaConfig     *config.WebDriver
	deviceStorage storage.Device

	refreshMu sync.Mutex // serializes refreshes
	mu        sync.Mutex // guards lastPoll and stopWatch
	lastPoll  time.Time
	stopWatch context.CancelFunc

	masterURL      string
	nodeIdentifier string
	authToken      *string

	notifyChange func() // set by the device manager, may be nil
	tracking     atomic.Bool
}

// NewHandler creates the handler with the go-ios and xcodebuild defaults. It
// adds the hook that reads the WebDriverAgent URL from the go-ios log output to
// the standard logger; use NewHandlerWithDeps to wire the hook yourself.
func NewHandler(cfg config.Manager, ds storage.Device) *Handler {
	hook := &webdriver.WDAConnectHook{}
	logrus.AddHook(hook)
	return NewHandlerWithDeps(cfg, ds, Deps{Hook: hook})
}

// NewHandlerWithDeps creates the handler with its collaborators, see Deps. It
// doesn't register the hook with a logger; add Hook() to it.
func NewHandlerWithDeps(cfg config.Manager, ds storage.Device, deps Deps) *Handler {
	return &Handler{deps: deps.withDefaults(), deviceStorage: ds, wdaConfig: cfg.WebDriver}
}

// Hook is the log hook that reads the WebDriverAgent URL of go-ios. It must be
// added to the logger the go-ios library logs to, e.g. logrus.AddHook.
func (m *Handler) Hook() *webdriver.WDAConnectHook {
	return m.deps.Hook
}

// newDevice creates a device that starts WebDriverAgent with the manager's
// config.
func (m *Handler) newDevice(id string) *Device {
	return newDevice(id, m.wdaConfig, m.deps)
}

// SetChangeNotifier implements device.ChangeNotifying.
func (m *Handler) SetChangeNotifier(notify func()) {
	m.notifyChange = notify
}

func (m *Handler) changed() {
	if m.notifyChange != nil {
		m.notifyChange()
	}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	m.masterURL = masterUrl
	m.nodeIdentifier = nodeIdentifier
	m.authToken = authToken
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for i := range devs {
		dev := m.newDevice(devs[i].DeviceIdentifier)
		dev.SetConfig(devs[i])
		m.wireRemote(dev)
		m.Add(dev)
	}
	return nil
}

// wireRemote sends the logs and app actions of the device to the master.
func (m *Handler) wireRemote(d *Device) {
	d.SetLogWriter(generic.NewRemoteLogWriter(m.masterURL, m.nodeIdentifier, d.ID, m.authToken))
	d.AddActionHandler(remote.NewActionHandler(m.masterURL, m.nodeIdentifier, d.ID, m.authToken))
}

// StartDevice starts WebDriverAgent on the device.
func (m *Handler) StartDevice(deviceID string) error {
	return m.withDevice(deviceID, (*Device).StartXCUITestRunner)
}

// StopDevice ends WebDriverAgent on the device.
func (m *Handler) StopDevice(deviceID string) error {
	return m.withDevice(deviceID, (*Device).StopXCUITestRunner)
}

func (m *Handler) withDevice(deviceID string, fn func(*Device) error) error {
	dev, ok := m.Get(deviceID)
	if !ok {
		return device.DeviceNotFoundError
	}
	return fn(dev)
}

func (m *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device: %w", device.ErrUnsupported)
}
