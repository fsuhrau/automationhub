package manager

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// NodeRegistry knows the nodes and their connection state.
type NodeRegistry interface {
	Run(ctx context.Context)
	IsNodeKnown(nodeIdentifier NodeIdentifier) bool
	GetStatus(nodeIdentifier NodeIdentifier) (int, error)
	RegisterNode(nodeIdentifier NodeIdentifier, client RPCClient, hostname, address, operationSystem, environment string, port int32, managers []string)
	GetNodes() []NodeIdentifier
	GetManagers(nodeIdentifier NodeIdentifier) (map[string][]device.Device, error)
}

// NodeLifecycle starts and stops the devices a node manages.
type NodeLifecycle interface {
	StartDevice(nodeIdentifier NodeIdentifier, deviceId string) error
	StopDevice(nodeIdentifier NodeIdentifier, deviceId string) error
}

// NodeApps uploads, installs and runs apps on the devices of a node.
type NodeApps interface {
	IsAppUploaded(node NodeIdentifier, parameter *app.Parameter) (bool, error)
	UploadApp(ctx context.Context, node NodeIdentifier, parameter *app.Parameter) error
	IsAppInstalled(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) (bool, error)
	InstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	UninstallApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	StartApp(nodeIdentifier NodeIdentifier, deviceId string, config *device.DeviceConfig, parameter *app.Parameter, sessionId string, nodeUrl string) error
	StopApp(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
	ClearAppData(nodeIdentifier NodeIdentifier, deviceId string, parameter *app.Parameter) error
}

// NodeConnection reports the app connection of a node device.
type NodeConnection interface {
	IsConnected(nodeIdentifier NodeIdentifier, deviceId string) bool
	ConnectionTimeout(nodeIdentifier NodeIdentifier, deviceId string) time.Duration
}

// NodeControl drives a node device: native control and app actions.
type NodeControl interface {
	Control(nodeIdentifier NodeIdentifier, req device.ControlRequest) (device.ControlResponse, error)
	SendAction(nodeIdentifier NodeIdentifier, deviceId string, action []byte)
}

// NodeMedia streams the video of a node device.
type NodeMedia interface {
	StartVideoStream(nodeIdentifier NodeIdentifier, req device.VideoStreamRequest) error
}

// Nodes is the master's view of all nodes. Consumers that need only one role
// should depend on that role's interface.
type Nodes interface {
	NodeRegistry
	NodeLifecycle
	NodeApps
	NodeConnection
	NodeControl
	NodeMedia
}

// The optional per-device roles of a node. Not every node connection offers
// them, so they are not part of Nodes: consumers assert them, see NodeDevice.

// NodeRecorder records the screen of a node device.
type NodeRecorder interface {
	StartRecording(nodeIdentifier NodeIdentifier, deviceId string, path string) error
	StopRecording(nodeIdentifier NodeIdentifier, deviceId string) error
}

// NodeScreenshotter takes the screenshot of a node device.
type NodeScreenshotter interface {
	GetScreenshot(nodeIdentifier NodeIdentifier, deviceId string) ([]byte, int, int, error)
}

// NodeFeatures runs the device specific features of a node device.
type NodeFeatures interface {
	HasFeature(nodeIdentifier NodeIdentifier, deviceId string, feature string) bool
	Execute(nodeIdentifier NodeIdentifier, deviceId string, data string)
}

// NodeNativeScripter runs native scripts on a node device.
type NodeNativeScripter interface {
	RunNativeScript(nodeIdentifier NodeIdentifier, deviceId string, script []byte)
}
