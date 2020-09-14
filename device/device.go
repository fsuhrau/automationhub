package device

import (
	"github.com/fsuhrau/automationhub/app"
	"net"
	"time"
)

type Device interface {
	DeviceID() string
	DeviceIP() net.IP
	DeviceState() State
	DeviceOSName() string
	DeviceOSVersion() string
	DeviceName() string

	IsAppInstalled(*app.Parameter) (bool, error)
	InstallApp(*app.Parameter) error
	UninstallApp(params *app.Parameter) error
	UpdateDeviceInfos() error

	ConnectionTimeout() time.Duration
	SetConnectionState(ConnectionState)
	IsAppConnected() bool
	StartApp(*app.Parameter, string, net.IP) error
	StopApp(*app.Parameter) error

	StartRecording(string) error
	StopRecording() error

	GetScreenshot() ([]byte, int, int, error)

	HasFeature(string) bool
	Execute(string)
}
