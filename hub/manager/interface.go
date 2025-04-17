package manager

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/sirupsen/logrus"
	"time"
)

type NodeIdentifier string

type Devices interface {
	Run(ctx context.Context, runSocketListener bool) error
	Devices() (map[string][]device.Device, error)
	SendAction(dev device.Device, a action.Interface) error
	// UnlockDevice(s Session) error
	// LockDevice(s Session, properties *device.Properties) (*DeviceLock, error)
	//GetDevice(id uint) (device.Device, error)
	GetDevice(id string) (device.Device, string)
	Start(getDevice device.Device) error
	Stop(getDevice device.Device) error
	AddHandler(handler device.Handler)
	StopObserver()
	FindDevice(params *app.Parameter) device.Device
	RegisterDevice(data device.RegisterData) (device.Device, error)
	Managers() []string
}

type Sessions interface {
	Run(ctx context.Context)
	GetSession(sessionID string) (Session, error)
	CreateNewSession(logger *logrus.Entry, properties *device.Properties, parameter *app.Parameter) Session
	AddSession(s Session)
	StopSession(s Session) error
	GetSessions() []Session
	GetSessionDetails(sessionID string) Session
}

type Storage interface {
	StoreSceneGraph(data []byte) (string, error)
	StoreImage(content []byte) (string, error)
	Close() error
}

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

	// Device Actions
	IsAppUploaded(node NodeIdentifier, parameter *app.Parameter) (bool, error)
	UploadApp(ctx context.Context, node NodeIdentifier, parameter *app.Parameter) error
	IsAppInstalled(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) (bool, error)
	InstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	UninstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	StartApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
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
