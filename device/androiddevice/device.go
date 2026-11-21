package androiddevice

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/adb"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
)

var (
	// DeviceLockRegex reads mDreamingLockscreen from "dumpsys window".
	DeviceLockRegex = regexp.MustCompile(`.*mDreamingLockscreen=(true|false).*`)
	// DeviceAwakeRegex reads mHoldingDisplaySuspendBlocker from "dumpsys power".
	DeviceAwakeRegex = regexp.MustCompile(`.*mHoldingDisplaySuspendBlocker=(true|false).*`)
)

// ConnectionTimeout is how long the app may take to connect after it started.
const ConnectionTimeout = 2 * time.Minute

// Compile-time checks of the roles an Android device plays.
var (
	_ device.Device         = (*Device)(nil)
	_ device.AppDataClearer = (*Device)(nil)
	_ device.InputCapable   = (*Device)(nil)
	_ device.LogReader      = (*Device)(nil)
	_ device.LogFollower    = (*Device)(nil)
	_ device.VideoStreamer  = (*Device)(nil)
)

// Device is an Android phone or emulator reached through adb.
type Device struct {
	generic.Device
	generic.Info
	deps      Deps
	video     sessionStarter
	deviceIP  net.IP
	apiLevel  int64
	recording recording
	// installedApps maps an app identifier to the hash of the installed build
	installedApps map[string]string
}

func newDevice(deps Deps, video sessionStarter, id string) *Device {
	return &Device{
		Info: generic.Info{
			ID:       id,
			Name:     id,
			OSName:   "android",
			Type:     int(models.DeviceTypePhone),
			Platform: int(models.PlatformTypeAndroid),
		},
		deps:          deps,
		video:         video,
		installedApps: map[string]string{},
	}
}

// adb returns the adb client of the device.
func (d *Device) adb() adb.Client {
	return adb.Client{Runner: d.deps.Runner, Serial: d.ID}
}

// run runs "adb -s <id> <args>" and returns its stdout.
func (d *Device) run(args ...string) ([]byte, error) {
	return d.adb().Run(context.Background(), args...)
}

// DeviceIP is the address of the device in the local network.
func (d *Device) DeviceIP() net.IP { return d.deviceIP }

// SetDeviceState sets the state by its wire name, e.g. "StateBooted".
func (d *Device) SetDeviceState(state string) { d.State = device.ParseState(state) }

// GetConnectionString is the adb address of a network device.
func GetConnectionString(parameter *models.ConnectionParameter) string {
	return fmt.Sprintf("%s:%d", parameter.IP, parameter.Port)
}

func (d *Device) IsAppConnected() bool { return d.Connection() != nil }

func (d *Device) ConnectionTimeout() time.Duration { return ConnectionTimeout }

// HasFeature implements device.FeatureExecutor.
func (d *Device) HasFeature(feature string) bool { return feature == featureBack }

const featureBack = "back"

// Execute implements device.FeatureExecutor.
func (d *Device) Execute(feature string) {
	d.Log("device", "Execute Feature: '%s'", feature)
	if feature == featureBack {
		_ = d.pressKey(KEYCODE_BACK)
	}
}

// RunNativeScript implements device.NativeScripter.
func (d *Device) RunNativeScript(script []byte) {
	if err := New(d).Execute(string(script)); err != nil {
		d.Error("device", "Script Failed: %v", err)
		_ = d.pressKey(KEYCODE_BACK)
	}
}
