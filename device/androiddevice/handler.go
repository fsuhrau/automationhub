package androiddevice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/adb"
	"github.com/fsuhrau/automationhub/device/androiddevice/scrcpy"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/remote"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

const (
	Manager = "android_device"
)

// fallbackPollInterval: while the adb event stream is connected, the device
// list is still polled at this interval as a safety net.
const fallbackPollInterval = time.Minute

// Handler manages the Android devices of the node.
type Handler struct {
	generic.Registry[*Device]
	deps          Deps
	video         sessionStarter
	deviceStorage storage.Device

	mu             sync.Mutex // guards lastPoll, stopWatch and the device list updates
	masterURL      string
	nodeIdentifier string
	authToken      *string
	adbServerAddr  string
	notifyChange   func()      // set by the device manager, may be nil
	tracking       atomic.Bool // the adb track-devices stream is connected
	lastPoll       time.Time
	stopWatch      context.CancelFunc
}

// NewHandler creates the handler with the real adb and the system clock.
func NewHandler(ds storage.Device) *Handler {
	return NewHandlerWithDeps(ds, Deps{})
}

// NewHandlerWithDeps creates the handler with the given dependencies; nil
// fields use the defaults.
func NewHandlerWithDeps(ds storage.Device, deps Deps) *Handler {
	deps = deps.withDefaults()
	launcher := scrcpy.NewLauncher(scrcpy.Deps{Runner: deps.Runner, Clock: deps.Clock, SessionID: deps.SessionID})
	return &Handler{
		deps:          deps,
		video:         scrcpyStarter{launcher},
		deviceStorage: ds,
		adbServerAddr: DefaultADBServerAddr,
	}
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

// adbHost runs an adb command that is not bound to a device.
func (m *Handler) adbHost(args ...string) ([]byte, error) {
	return m.deps.Runner.Run(context.Background(), adb.Binary, args...)
}

// Init restores the devices from storage, lists the devices adb sees and
// disconnects the network devices again.
func (m *Handler) Init(masterURL, nodeIdentifier string, authToken *string) error {
	m.authToken = authToken
	m.masterURL = masterURL
	m.nodeIdentifier = nodeIdentifier
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for _, cfg := range devs {
		if err := m.restoreDevice(cfg); err != nil {
			return err
		}
	}
	if err := m.RefreshDevices(true); err != nil {
		return err
	}
	for _, cfg := range devs {
		if err := m.adbRemote("disconnect", cfg); err != nil {
			return err
		}
	}
	return nil
}

// restoreDevice adds a stored device and connects it when it is a network device.
func (m *Handler) restoreDevice(cfg *models.Device) error {
	dev := newDevice(m.deps, m.video, cfg.DeviceIdentifier)
	dev.SetConfig(cfg)
	m.wireRemote(dev)
	m.Add(dev)
	return m.adbRemote("connect", cfg)
}

// adbRemote runs "adb connect|disconnect <ip:port>" for network devices; other
// devices are left alone.
func (m *Handler) adbRemote(command string, cfg *models.Device) error {
	if cfg == nil || cfg.ConnectionParameter == nil || cfg.ConnectionParameter.ConnectionType != models.ConnectionTypeRemote {
		return nil
	}
	_, err := m.adbHost(command, GetConnectionString(cfg.ConnectionParameter))
	return err
}

// wireRemote sends the log and the app actions of the device to the master.
func (m *Handler) wireRemote(d *Device) {
	d.SetLogWriter(generic.NewRemoteLogWriter(m.masterURL, m.nodeIdentifier, d.ID, m.authToken))
	d.AddActionHandler(remote.NewActionHandler(m.masterURL, m.nodeIdentifier, d.ID, m.authToken))
}

// RegisterDevice implements device.Handler; Android devices are found by adb.
func (m *Handler) RegisterDevice(device.RegisterData) (device.Device, error) {
	return nil, fmt.Errorf("register device: %w", device.ErrUnsupported)
}

// persist stores the device state; a failing store doesn't stop the refresh.
func (m *Handler) persist(d *Device) {
	if err := m.deviceStorage.Update(m.Name(), d); err != nil {
		logrus.Warnf("store android device %s: %v", d.ID, err)
	}
}
