package device

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device/appproto"
)

// Identity describes what a device is.
type Identity interface {
	DeviceID() string
	DeviceState() State
	DeviceOSName() string
	DeviceOSVersion() string
	DeviceName() string
	DeviceParameter() map[string]string
	DeviceType() int
	PlatformType() int
	TargetVersion() string
}

// AppManager installs, starts and stops the app under test.
type AppManager interface {
	IsAppInstalled(*app.Parameter) (bool, error)
	InstallApp(*app.Parameter) error
	UninstallApp(*app.Parameter) error
	StartApp(deviceConfig *DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(*app.Parameter) error
}

// AppConnection is the link to the running app: its context, its websocket
// connection and sending data to it.
type AppConnection interface {
	NewContext() context.Context
	Cancel()
	ConnectionTimeout() time.Duration
	SetConnection(*Connection)
	Connection() *Connection
	IsAppConnected() bool
	Send([]byte) error
}

// Locker reserves a device for one test run.
type Locker interface {
	Lock() error
	Unlock() error
	IsLocked() bool
}

// TestLogger collects the log, data and result of a test run.
type TestLogger interface {
	SetLogWriter(LogWriter)
	GetLogWriter() LogWriter
	Data(string, string)
	SetKey(key, value string)
	LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string)
	Log(string, string, ...interface{})
	Error(string, string, ...interface{})
	Exception(string, string, ...interface{})
	Passed(result bool)
}

// Recorder records the screen of a device.
type Recorder interface {
	StartRecording(string) error
	StopRecording() error
}

// Screenshotter takes a screenshot and returns it with its width and height.
type Screenshotter interface {
	GetScreenshot() ([]byte, int, int, error)
}

// FeatureExecutor runs device specific features.
type FeatureExecutor interface {
	HasFeature(string) bool
	Execute(string)
}

// NativeScripter runs a native script on the device.
type NativeScripter interface {
	RunNativeScript(data []byte)
}

// ActionHandlerRegistry holds the handlers that get the app's responses.
type ActionHandlerRegistry interface {
	AddActionHandler(appproto.ActionHandler)
	RemoveActionHandler(appproto.ActionHandler)
	ActionHandlers() []appproto.ActionHandler
}
