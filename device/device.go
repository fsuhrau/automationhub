package device

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/action"
	"time"
)

type Device interface {
	DeviceID() string
	DeviceState() State
	DeviceOSName() string
	DeviceOSVersion() string
	DeviceName() string
	DeviceModel() string

	// IsAppBundleAvailable(*app.Parameter) (bool, error)
	IsAppInstalled(*app.Parameter) (bool, error)
	InstallApp(*app.Parameter) error
	UninstallApp(*app.Parameter) error
	UpdateDeviceInfos() error

	NewContext() context.Context
	Cancel()
	ConnectionTimeout() time.Duration
	SetConnection(*Connection)
	Connection() *Connection
	IsAppConnected() bool

	StartApp(*app.Parameter, string, string) error
	StopApp(*app.Parameter) error
	Send([]byte) error

	StartRecording(string) error
	StopRecording() error

	GetScreenshot() ([]byte, int, int, error)

	HasFeature(string) bool
	Execute(string)
	Lock() error
	Unlock() error
	IsLocked() bool

	SetLogWriter(LogWriter)
	GetLogWriter() LogWriter
	Data(string, string)
	LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string)
	Log(string, string, ...interface{})
	Error(string, string, ...interface{})
	Exception(string, string, ...interface{})
	AddActionHandler(action.ActionHandler)
	RemoveActionHandler(action.ActionHandler)
	ActionHandlers() []action.ActionHandler
	RunNativeScript(data []byte)
}
