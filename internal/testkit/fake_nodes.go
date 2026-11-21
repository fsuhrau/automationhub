package testkit

import (
	"context"
	"slices"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

var _ manager.Nodes = (*FakeNodes)(nil)

// FakeNodes is a manager.Nodes that records its calls. The exported fields
// are the answers of the query methods; FailWith scripts errors.
type FakeNodes struct {
	recorder

	Known        []manager.NodeIdentifier
	Status       int
	Managed      map[string][]device.Device
	Uploaded     bool
	Installed    bool
	Connected    bool
	Feature      bool
	ControlReply device.ControlResponse
	PNG          []byte
	Width        int
	Height       int
	Timeout      time.Duration
}

func (n *FakeNodes) Run(context.Context) { _ = n.record("Run") }

func (n *FakeNodes) IsNodeKnown(id manager.NodeIdentifier) bool {
	_ = n.record("IsNodeKnown", id)
	n.mu.Lock()
	defer n.mu.Unlock()
	return slices.Contains(n.Known, id)
}

func (n *FakeNodes) GetStatus(id manager.NodeIdentifier) (int, error) {
	return n.Status, n.record("GetStatus", id)
}

//nolint:revive // implements manager.Nodes
func (n *FakeNodes) RegisterNode(id manager.NodeIdentifier, client manager.RPCClient, hostname, address, operationSystem, environment string, port int32, managers []string) {
	_ = n.record("RegisterNode", id, client, hostname, address, operationSystem, environment, port, managers)
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Known = append(n.Known, id)
}

func (n *FakeNodes) GetNodes() []manager.NodeIdentifier {
	_ = n.record("GetNodes")
	n.mu.Lock()
	defer n.mu.Unlock()
	return slices.Clone(n.Known)
}

func (n *FakeNodes) GetManagers(id manager.NodeIdentifier) (map[string][]device.Device, error) {
	return n.Managed, n.record("GetManagers", id)
}

func (n *FakeNodes) StartDevice(id manager.NodeIdentifier, deviceID string) error {
	return n.record("StartDevice", id, deviceID)
}

func (n *FakeNodes) StopDevice(id manager.NodeIdentifier, deviceID string) error {
	return n.record("StopDevice", id, deviceID)
}

func (n *FakeNodes) IsAppUploaded(id manager.NodeIdentifier, p *app.Parameter) (bool, error) {
	return n.Uploaded, n.record("IsAppUploaded", id, p)
}

func (n *FakeNodes) UploadApp(_ context.Context, id manager.NodeIdentifier, p *app.Parameter) error {
	return n.record("UploadApp", id, p)
}

func (n *FakeNodes) IsAppInstalled(id manager.NodeIdentifier, deviceID string, p *app.Parameter) (bool, error) {
	return n.Installed, n.record("IsAppInstalled", id, deviceID, p)
}

func (n *FakeNodes) InstallApp(id manager.NodeIdentifier, deviceID string, p *app.Parameter) error {
	return n.record("InstallApp", id, deviceID, p)
}

func (n *FakeNodes) UninstallApp(id manager.NodeIdentifier, deviceID string, p *app.Parameter) error {
	return n.record("UninstallApp", id, deviceID, p)
}

//nolint:revive // implements manager.Nodes
func (n *FakeNodes) StartApp(id manager.NodeIdentifier, deviceID string, cfg *device.DeviceConfig, p *app.Parameter, sessionID, nodeURL string) error {
	return n.record("StartApp", id, deviceID, cfg, p, sessionID, nodeURL)
}

func (n *FakeNodes) StopApp(id manager.NodeIdentifier, deviceID string, p *app.Parameter) error {
	return n.record("StopApp", id, deviceID, p)
}

func (n *FakeNodes) ClearAppData(id manager.NodeIdentifier, deviceID string, p *app.Parameter) error {
	return n.record("ClearAppData", id, deviceID, p)
}

func (n *FakeNodes) Control(id manager.NodeIdentifier, req device.ControlRequest) (device.ControlResponse, error) {
	return n.ControlReply, n.record("Control", id, req)
}

func (n *FakeNodes) StartVideoStream(id manager.NodeIdentifier, req device.VideoStreamRequest) error {
	return n.record("StartVideoStream", id, req)
}

func (n *FakeNodes) IsConnected(id manager.NodeIdentifier, deviceID string) bool {
	_ = n.record("IsConnected", id, deviceID)
	return n.Connected
}

func (n *FakeNodes) StartRecording(id manager.NodeIdentifier, deviceID, path string) error {
	return n.record("StartRecording", id, deviceID, path)
}

func (n *FakeNodes) StopRecording(id manager.NodeIdentifier, deviceID string) error {
	return n.record("StopRecording", id, deviceID)
}

func (n *FakeNodes) GetScreenshot(id manager.NodeIdentifier, deviceID string) ([]byte, int, int, error) {
	return n.PNG, n.Width, n.Height, n.record("GetScreenshot", id, deviceID)
}

func (n *FakeNodes) HasFeature(id manager.NodeIdentifier, deviceID, feature string) bool {
	_ = n.record("HasFeature", id, deviceID, feature)
	return n.Feature
}

func (n *FakeNodes) Execute(id manager.NodeIdentifier, deviceID, data string) {
	_ = n.record("Execute", id, deviceID, data)
}

func (n *FakeNodes) ConnectionTimeout(id manager.NodeIdentifier, deviceID string) time.Duration {
	_ = n.record("ConnectionTimeout", id, deviceID)
	return n.Timeout
}

func (n *FakeNodes) RunNativeScript(id manager.NodeIdentifier, deviceID string, script []byte) {
	_ = n.record("RunNativeScript", id, deviceID, script)
}

func (n *FakeNodes) SendAction(id manager.NodeIdentifier, deviceID string, a []byte) {
	_ = n.record("SendAction", id, deviceID, a)
}
