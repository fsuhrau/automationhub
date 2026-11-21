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

// Device is the composite of every role a device can play. G33 removes the
// optional roles from it; consumers then reach them with As.
type Device interface {
	Identity
	AppManager
	AppConnection
	Locker
	TestLogger

	Recorder
	Screenshotter
	FeatureExecutor
	NativeScripter
	ActionHandlerRegistry
}
