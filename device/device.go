package device

import (
	"errors"
	"github.com/fsuhrau/automationhub/app"
)

type DeviceConfig struct {
	Connection      int
	IP              string
	Port            int
	DeviceParameter map[string]string
}

func GetAttribute(config *DeviceConfig, attribute string) string {
	if v, ok := config.DeviceParameter[attribute]; ok {
		return v
	}
	return ""
}

// ErrUnsupported is returned by optional device capabilities that the
// platform does not implement.
var ErrUnsupported = errors.New("operation not supported by this device")

// AppDataClearer is implemented by devices that can wipe an app's data.
type AppDataClearer interface {
	ClearAppData(*app.Parameter) error
}

// Device is the composite of the roles every device plays. The optional
// roles (Recorder, Screenshotter, FeatureExecutor, NativeScripter, and the
// ones in video.go and control.go) are reached with As.
//
// ActionHandlerRegistry stays in the core: every platform gets it from
// generic.Device, and the device manager, the action executors and the app
// connection use it unconditionally on every device.
type Device interface {
	Identity
	AppManager
	AppConnection
	Locker
	TestLogger
	ActionHandlerRegistry
}
