package device

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/action"
	"time"
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

type Device interface {
	DeviceID() string
	DeviceState() State
	DeviceOSName() string
	DeviceOSVersion() string
	DeviceName() string
	DeviceParameter() map[string]string
	DeviceType() int
	PlatformType() int

	TargetVersion() string

	// IsAppBundleAvailable(*app.Parameter) (bool, error)
	IsAppInstalled(*app.Parameter) (bool, error)
	InstallApp(*app.Parameter) error
	UninstallApp(*app.Parameter) error

	NewContext() context.Context
	Cancel()
	ConnectionTimeout() time.Duration
	SetConnection(*Connection)
	Connection() *Connection
	IsAppConnected() bool

	StartApp(deviceConfig *DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error
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
	Passed(result bool)
}
