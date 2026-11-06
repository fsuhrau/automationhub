package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
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

func (d *NodeDevice) DeviceType() int {
	return d.deviceType
}

func (d *NodeDevice) PlatformType() int {
	return int(d.platformType)
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

func (d *NodeDevice) DeviceState() device.State {
	return d.deviceState
}

func (d *NodeDevice) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	case "StateRemoteDisconnected":
		d.deviceState = device.StateRemoteDisconnected
	case "StateNodeDisconnected":
		d.deviceState = device.StateNodeDisconnected
	default:
		d.deviceState = device.StateUnknown
	}
}

func (d *NodeDevice) UpdateDeviceInfos(parameter map[string]string) {
	d.deviceParameter = []models.DeviceParameter{}
	for k, v := range parameter {
		d.deviceParameter = append(d.deviceParameter, models.DeviceParameter{
			Key:   k,
			Value: v,
		})
	}
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

func (d *NodeDevice) StartApp(_ *device.DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error {

	var (
		connectionType int
		ip             string
		port           int
		params         map[string]string
	)
	params = make(map[string]string)

	if d.Device.Config.ConnectionParameter != nil {
		connectionType = int(d.Device.Config.ConnectionParameter.ConnectionType)
		ip = d.Device.Config.ConnectionParameter.IP
		port = d.Device.Config.ConnectionParameter.Port
	}
	if d.Device.Config.CustomParameter != nil {
		for _, param := range d.Device.Config.CustomParameter {
			params[param.Key] = param.Value
		}
	}

	deviceConfig := device.DeviceConfig{
		Connection:      connectionType,
		IP:              ip,
		Port:            port,
		DeviceParameter: params,
	}

	return d.nodeManager.StartApp(d.nodeId, d.deviceID, &deviceConfig, appParams, sessionId, nodeUrl)
}

func (d *NodeDevice) StopApp(params *app.Parameter) error {
	return d.nodeManager.StopApp(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) ClearAppData(params *app.Parameter) error {
	return d.nodeManager.ClearAppData(d.nodeId, d.deviceID, params)
}

func (d *NodeDevice) control(req device.ControlRequest) (device.ControlResponse, error) {
	req.DeviceID = d.deviceID
	return d.nodeManager.Control(d.nodeId, req)
}

// Tap implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) Tap(x, y int) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlTap, X: x, Y: y})
	return err
}

// Swipe implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) Swipe(fromX, fromY, toX, toY int, duration time.Duration) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlSwipe, X: fromX, Y: fromY, ToX: toX, ToY: toY, DurationMs: int(duration.Milliseconds())})
	return err
}

// InputText implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) InputText(text string) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlText, Text: text})
	return err
}

// PressKey implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) PressKey(key string) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlKey, Key: key})
	return err
}

// ReadLogs implements device.LogReader by forwarding to the node.
func (d *NodeDevice) ReadLogs(lines int) ([]string, error) {
	resp, err := d.control(device.ControlRequest{Op: device.ControlReadLogs, Lines: lines})
	return resp.Lines, err
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
