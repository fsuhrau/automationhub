package devices

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

	IsAppInstalled(*app.Parameter) bool
	InstallApp(*app.Parameter) error
	UninstallApp(string) error
	UpdateParameter() error

	ConnectionTimeout() time.Duration
	SetConnectionState(ConnectionState)
	IsAppConnected() bool
	StartApp(*app.Parameter, string, net.IP) error
	StopApp(*app.Parameter) error

	StartRecording(string) error
	StopRecording() error
}
