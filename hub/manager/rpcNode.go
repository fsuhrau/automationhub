package manager

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// RPCLink checks and lists a node over its RPC connection.
type RPCLink interface {
	Ping() error
	GetDevices() (map[string][]device.Device, error)
}

// RPCLifecycle starts and stops devices on the node.
type RPCLifecycle interface {
	StartDevice(deviceId string) error
	StopDevice(deviceId string) error
}

// RPCApps uploads, installs and runs apps on the node.
type RPCApps interface {
	IsAppInstalled(deviceId string, parameter *app.Parameter) (bool, error)
	InstallApp(deviceId string, parameter *app.Parameter) error
	UninstallApp(deviceId string, parameter *app.Parameter) error
	StartApp(deviceId string, config *device.DeviceConfig, parameter *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(deviceId string, parameter *app.Parameter) error
	ClearAppData(deviceId string, parameter *app.Parameter) error
	UploadApp(ctx context.Context, parameter *app.Parameter) error
	IsAppUploaded(parameter *app.Parameter) (bool, error)
}

// RPCConnection reports the app connection of a device on the node.
type RPCConnection interface {
	IsConnected(deviceId string) bool
	ConnectionTimeout(deviceId string) time.Duration
}

// RPCControl drives a device on the node.
type RPCControl interface {
	Control(req device.ControlRequest) (device.ControlResponse, error)
	HasFeature(deviceId string, feature string) bool
	Execute(deviceId string, data string)
	RunNativeScript(deviceId string, script []byte)
	SendAction(deviceId string, action []byte)
}

// RPCMedia captures the screen of a device on the node.
type RPCMedia interface {
	StartVideoStream(req device.VideoStreamRequest) error
	StartRecording(deviceId string, path string) error
	StopRecording(deviceId string) error
	GetScreenshot(deviceId string) ([]byte, int, int, error)
}

// RPCClient is the master's RPC connection to one node.
type RPCClient interface {
	RPCLink
	RPCLifecycle
	RPCApps
	RPCConnection
	RPCControl
	RPCMedia
}
