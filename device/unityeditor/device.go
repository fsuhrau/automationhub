package unityeditor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
)

const ConnectionTimeout = 60 * time.Minute

// Device is a Unity editor that registered itself with the hub. It supports
// the Identity, AppManager (start and stop only), AppConnection, Locker,
// TestLogger and ActionHandlerRegistry roles; recording and screenshots are
// not supported.
type Device struct {
	generic.Device
	generic.Info

	clock device.Clock

	// mu guards the fields below.
	mu              sync.Mutex
	osInfos         string
	ip              net.IP
	conn            managerConn
	lastUpdateAt    time.Time
	updated         bool
	sendCh          chan []byte
	process         device.Process
	instanceLogFile string
	logCancel       context.CancelFunc
}

func newDevice(info generic.Info, clock device.Clock) *Device {
	info.Type = int(models.DeviceTypeUnityEditor)
	info.Platform = int(models.PlatformTypeEditor)
	return &Device{Info: info, clock: clock}
}

func (d *Device) DeviceIP() net.IP {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ip
}

func (d *Device) DeviceState() device.State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State
}

// SetDeviceState sets the state from its name, see device.ParseState.
func (d *Device) SetDeviceState(state string) {
	d.setState(device.ParseState(state))
}

func (d *Device) setState(state device.State) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State = state
}

// IsAppInstalled is always true: the editor runs the project, there is
// nothing to install.
func (d *Device) IsAppInstalled(_ *app.Parameter) (bool, error) {
	return true, nil
}

func (d *Device) InstallApp(_ *app.Parameter) error {
	return fmt.Errorf("install app: %w", device.ErrUnsupported)
}

func (d *Device) UninstallApp(_ *app.Parameter) error {
	return fmt.Errorf("uninstall app: %w", device.ErrUnsupported)
}

func (d *Device) StartRecording(string) error {
	return fmt.Errorf("start recording: %w", device.ErrUnsupported)
}

func (d *Device) StopRecording() error {
	return fmt.Errorf("stop recording: %w", device.ErrUnsupported)
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("screenshot: %w", device.ErrUnsupported)
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) HasFeature(string) bool {
	return false
}

// Execute has no features to run, see HasFeature.
func (d *Device) Execute(string) {}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

// RunNativeScript does nothing: the editor has no native scripts.
func (d *Device) RunNativeScript([]byte) {}
