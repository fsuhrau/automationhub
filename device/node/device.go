package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"net"
	"os/exec"
	"time"
)

const ConnectionTimeout = 10 * time.Second

type NodeDevice struct {
	generic.Device
	deviceModel      string
	deviceOSName     string
	deviceOSVersion  string
	deviceName       string
	deviceID         string
	deviceType       int
	targetVersion    string
	platformType     models.PlatformType
	deviceIP         net.IP
	deviceState      device.State
	recordingSession *exec.Cmd
	lastUpdateAt     time.Time

	nodeId          manager.NodeIdentifier
	nodeManager     manager.Nodes
	deviceParameter []models.DeviceParameter
}

func (d *NodeDevice) DeviceParameter() map[string]string {
	parameter := make(map[string]string)
	for _, p := range d.deviceParameter {
		parameter[p.Key] = p.Value
	}
	return parameter
}

func (d *NodeDevice) GetNodeID() manager.NodeIdentifier {
	return d.nodeId
}

func (d *NodeDevice) DeviceModel() string {
	return d.deviceModel
}

func (d *NodeDevice) DeviceType() int {
	return d.deviceType
}

func (d *NodeDevice) PlatformType() int {
	return int(d.platformType)
}

func (d *NodeDevice) Parameter() string {
	return ""
}

func (d *NodeDevice) DeviceOSName() string {
	return d.deviceOSName
}

func (d *NodeDevice) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *NodeDevice) TargetVersion() string {
	return d.targetVersion
}

func (d *NodeDevice) DeviceName() string {
	return d.deviceName
}

func (d *NodeDevice) DeviceID() string {
	return d.deviceID
}

func (d *NodeDevice) DeviceIP() net.IP {
	return d.deviceIP
}

func (d *NodeDevice) DeviceState() device.State {
	return d.deviceState
}

func (d *NodeDevice) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *NodeDevice) UpdateDeviceInfosFromParameter(parameter string) {
	switch d.platformType {
	case models.PlatformTypeAndroid:
		d.deviceParameter = androiddevice.UnpackDeviceParameter(parameter)
	case models.PlatformTypeMac:
		d.deviceParameter = macos.UnpackDeviceParameter(parameter)
	}
}

func (d *NodeDevice) UpdateDeviceInfos() error {
	return nil
}

func (d *NodeDevice) IsAppInstalled(params *app.Parameter) (bool, error) {
	return d.nodeManager.IsAppInstalled(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) InstallApp(params *app.Parameter) error {
	return d.nodeManager.InstallApp(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) UninstallApp(params *app.Parameter) error {
	return d.nodeManager.UninstallApp(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) StartApp(params *app.Parameter, sessionId string, nodeUrl string) error {
	return d.nodeManager.StartApp(d.nodeId, d.deviceID, params, sessionId, nodeUrl)
}

func (d *NodeDevice) StopApp(params *app.Parameter) error {
	return d.nodeManager.StopApp(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) IsAppConnected() bool {
	// return d.Connection() != nil && d.Connection().Connection != nil
	return d.nodeManager.IsConnected(d.nodeId, d.deviceID)
}

func (d *NodeDevice) StartRecording(path string) error {
	return d.nodeManager.StartRecording(d.nodeId, d.deviceID, path)
}

func (d *NodeDevice) StopRecording() error {
	return d.nodeManager.StopRecording(d.nodeId, d.deviceID)
}

func (d *NodeDevice) GetScreenshot() ([]byte, int, int, error) {
	return d.nodeManager.GetScreenshot(d.nodeId, d.deviceID)
}

func (d *NodeDevice) HasFeature(feature string) bool {
	return d.nodeManager.HasFeature(d.nodeId, d.deviceID, feature)
}

func (d *NodeDevice) Execute(data string) {
	d.nodeManager.Execute(d.nodeId, d.deviceID, data)
}

func (d *NodeDevice) ConnectionTimeout() time.Duration {
	return d.nodeManager.ConnectionTimeout(d.nodeId, d.deviceID)
}

func (d *NodeDevice) RunNativeScript(script []byte) {
	d.nodeManager.RunNativeScript(d.nodeId, d.deviceID, script)
}

func (d *NodeDevice) Send(script []byte) error {
	d.nodeManager.SendAction(d.nodeId, d.deviceID, script)
	return nil
}

func (d *NodeDevice) NodeManager() manager.Nodes {
	return d.nodeManager
}
