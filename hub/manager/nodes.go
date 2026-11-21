package manager

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

type Nodes interface {
	Run(ctx context.Context)
	IsNodeKnown(nodeIdentifier NodeIdentifier) bool
	GetStatus(nodeIdentifier NodeIdentifier) (int, error)
	RegisterNode(nodeIdentifier NodeIdentifier, client RPCClient, hostname, address, operationSystem, environment string, port int32, managers []string)
	GetNodes() []NodeIdentifier
	GetManagers(nodeIdentifier NodeIdentifier) (map[string][]device.Device, error)

	// Manager Actions
	StartDevice(nodeIdentifier NodeIdentifier, deviceId string) error
	StopDevice(nodeIdentifier NodeIdentifier, deviceId string) error

	// device Actions
	IsAppUploaded(node NodeIdentifier, parameter *app.Parameter) (bool, error)
	UploadApp(ctx context.Context, node NodeIdentifier, parameter *app.Parameter) error
	IsAppInstalled(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) (bool, error)
	InstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	UninstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	StartApp(nodeIdentifier NodeIdentifier, deviceId string, config *device.DeviceConfig, parameter *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	ClearAppData(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	Control(nodeIdentifier NodeIdentifier, req device.ControlRequest) (device.ControlResponse, error)
	StartVideoStream(nodeIdentifier NodeIdentifier, req device.VideoStreamRequest) error
	IsConnected(nodeIdentifier NodeIdentifier, deviceId string) bool
	StartRecording(nodeIdentifier NodeIdentifier, deviceId string, path string) error
	StopRecording(nodeIdentifier NodeIdentifier, deviceId string) error
	GetScreenshot(nodeIdentifier NodeIdentifier, deviceId string) ([]byte, int, int, error)
	HasFeature(nodeIdentifier NodeIdentifier, deviceId string, feature string) bool
	Execute(nodeIdentifier NodeIdentifier, deviceId string, data string)
	ConnectionTimeout(nodeIdentifier NodeIdentifier, deviceId string) time.Duration
	RunNativeScript(nodeIdentifier NodeIdentifier, deviceId string, script []byte)
	SendAction(nodeIdentifier NodeIdentifier, deviceId string, action []byte)
}
