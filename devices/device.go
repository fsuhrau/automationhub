package devices

import (
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

	IsAppInstalled(string) bool
	InstallApp(string) error
	UninstallApp(string) error
	ExtractAppParameters(string) error

	ConnectionTimeout() time.Duration
	SetConnectionState(ConnectionState)
	IsAppConnected() bool
	StartApp(string, string, string, net.IP) error
	StopApp(string, string) error

	StartRecording(string) error
	StopRecording() error
}
