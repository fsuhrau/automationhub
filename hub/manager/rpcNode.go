package manager

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"time"
)

type RPCClient interface {
	Ping() error
	GetDevices() (map[string][]device.Device, error)
	StartDevice(deviceId string) error
	StopDevice(deviceId string) error
	IsAppInstalled(deviceId string, parameter *app.Parameter) (bool, error)
	InstallApp(deviceId string, parameter *app.Parameter) error
	UninstallApp(deviceId string, parameter *app.Parameter) error
	StartApp(deviceId string, parameter *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(deviceId string, parameter *app.Parameter) error
	IsConnected(deviceId string) bool
	StartRecording(deviceId string, path string) error
	StopRecording(deviceId string) error
	GetScreenshot(deviceId string) ([]byte, int, int, error)
	HasFeature(deviceId string, feature string) bool
	Execute(deviceId string, data string)
	ConnectionTimeout(deviceId string) time.Duration
	RunNativeScript(deviceId string, script []byte)
	SendAction(deviceId string, action []byte)
	UploadApp(ctx context.Context, parameter *app.Parameter) error
	IsAppUploaded(parameter *app.Parameter) (bool, error)
}
