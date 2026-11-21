package web

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/storage/models"
)

// ConnectionTimeout is how long a started page has to connect.
const ConnectionTimeout = 10 * time.Second

// Device is one browser installed on the node's machine. The embedded
// Info.Target holds the browser version.
type Device struct {
	generic.Device
	generic.Info
	env hostinfo.Env

	browser     string
	browserPath string
	updatedAt   time.Time
	process     device.Process
}

var _ device.Device = (*Device)(nil)

// browserConfig is a configured browser: its name and the path of its executable.
type browserConfig struct{ name, path string }

// newDevice reads the machine and browser infos for one configured browser.
func newDevice(ctx context.Context, env hostinfo.Env, b browserConfig) *Device {
	host := hostinfo.New(env)
	hostName, _ := host.HostName(ctx)
	info := generic.Info{
		Type:      int(models.DeviceTypeBrowser),
		Platform:  int(models.PlatformTypeWeb),
		State:     device.StateBooted,
		Parameter: host.Parameters(ctx, hostinfo.ParamModel, hostinfo.ParamSerial),
	}
	uuid, _ := host.HardwareUUID(ctx)
	info.ID = uuid + "/" + b.name
	info.OSName, _ = host.OSName(ctx)
	info.OSVersion, _ = host.OSVersion(ctx)
	d := &Device{Info: info, env: env, browser: b.name, browserPath: b.path, updatedAt: env.Clock.Now()}
	d.readBrowserVersion(ctx)
	d.Name = displayName(b.name, d.Target, hostName)
	return d
}

// displayName is "<browser> (<version>) @ <host>", without the version if unknown.
func displayName(browser, version, host string) string {
	if len(version) > 1 {
		return fmt.Sprintf("%s (%s) @ %s", browser, version, host)
	}
	return fmt.Sprintf("%s @ %s", browser, host)
}

// UpdatedAt is when the device infos were read.
func (d *Device) UpdatedAt() time.Time { return d.updatedAt }

func (d *Device) IsAppConnected() bool { return d.Connection() != nil }

func (d *Device) ConnectionTimeout() time.Duration { return ConnectionTimeout }

// HasFeature: a browser has no optional device features.
func (d *Device) HasFeature(string) bool { return false }

// Execute does nothing, there are no features to execute.
func (d *Device) Execute(string) {}

// RunNativeScript does nothing, a browser has no native script runtime.
func (d *Device) RunNativeScript([]byte) {}

func (d *Device) StartRecording(string) error {
	return fmt.Errorf("recording: %w", device.ErrUnsupported)
}

func (d *Device) StopRecording() error {
	return fmt.Errorf("recording: %w", device.ErrUnsupported)
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, fmt.Errorf("screenshot: %w", device.ErrUnsupported)
}
