package macos

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/storage/models"
)

// ConnectionTimeout is how long a started app has to connect.
const ConnectionTimeout = 10 * time.Second

// Device is the Mac the node runs on.
type Device struct {
	generic.Device
	generic.Info
	deps Deps

	updatedAt         time.Time
	runningExecutable string
}

var _ device.Device = (*Device)(nil)

// newDevice reads the identity and the hardware parameters of this Mac.
func newDevice(ctx context.Context, deps Deps) *Device {
	host := hostinfo.New(deps.Env)
	info := generic.Info{
		Type:      int(models.DeviceTypeDesktop),
		Platform:  int(models.PlatformTypeMac),
		State:     device.StateBooted,
		Parameter: host.Parameters(ctx, hardwareKeys()...),
	}
	info.ID, _ = host.HardwareUUID(ctx)
	info.OSName, _ = host.OSName(ctx)
	info.Name, _ = host.HostName(ctx)
	info.OSVersion, _ = host.OSVersion(ctx)
	return &Device{Info: info, deps: deps, updatedAt: deps.Clock.Now()}
}

func hardwareKeys() []string {
	return []string{
		hostinfo.ParamModel, hostinfo.ParamSerial, hostinfo.ParamCPU,
		hostinfo.ParamGPU, hostinfo.ParamRAM, hostinfo.ParamDriver,
	}
}

// UpdatedAt is when the device infos were read.
func (d *Device) UpdatedAt() time.Time { return d.updatedAt }

func (d *Device) IsAppConnected() bool { return d.Connection() != nil }

func (d *Device) ConnectionTimeout() time.Duration { return ConnectionTimeout }

// HasFeature: the Mac has no optional device features.
func (d *Device) HasFeature(string) bool { return false }

// Execute does nothing, there are no features to execute.
func (d *Device) Execute(string) {}

// RunNativeScript does nothing, the Mac has no native script runtime.
func (d *Device) RunNativeScript([]byte) {}
