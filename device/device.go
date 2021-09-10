package device

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/action"
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
	UninstallApp(*app.Parameter) error
	UpdateDeviceInfos() error

	ConnectionTimeout() time.Duration
	SetConnection(*Connection)
	Connection() *Connection
	IsAppConnected() bool

	StartApp(*app.Parameter, string, net.IP) error
	StopApp(*app.Parameter) error

	StartRecording(string) error
	StopRecording() error

	GetScreenshot() ([]byte, int, int, error)

	HasFeature(string) bool
	Execute(string)
	Lock() error
	Unlock() error
	IsLocker() bool

	SetLogWriter(LogWriter)
	Log(string, ...interface{})
	Error(string, ...interface{})
	ActionHandler() action.ActionHandler
	SetActionHandler(action.ActionHandler)
}
