package api

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
)

// DeviceListChannel is the SSE channel on which the UI learns that it has to
// reload the device list (a device was added or its details changed).
const DeviceListChannel = "device_list"

// NodeStatusPayload is published on the "nodes" SSE channel.
type NodeStatusPayload struct {
	NodeIdentifier manager.NodeIdentifier `json:"nodeIdentifier"`
	Connected      bool                   `json:"connected"`
}

// testRunChange is sent on the "test_runs" channel whenever a run started or
// finished, so pages that list runs (dashboard, tests, run history) refresh no
// matter whether the run was started from the UI, the CLI or MCP.
type testRunChange struct {
	ProjectID string `json:"projectId"`
	AppID     uint   `json:"appId"`
	TestID    uint   `json:"testId,omitempty"`
	TestRunID uint   `json:"testRunId"`
	State     string `json:"state"`
}

// registerSSEBridges publishes every event the UI listens to on its channel.
func registerSSEBridges(publisher sse.Publisher) {
	registerDeviceBridges(publisher)
	registerTestBridges(publisher)
	registerProtocolBridges(publisher)
	registerTestRunBridges(publisher)
	registerNodeBridges(publisher)
}

func registerDeviceBridges(p sse.Publisher) {
	events.DeviceStatusChanged.Register(newBridge(p, fixedChannel[events.DeviceStatusChangedPayload]("devices"), payloadContent))
	events.DeviceListChanged.Register(newBridge(p, fixedChannel[events.DeviceListChangedPayload](DeviceListChannel), payloadContent))
}

func registerTestBridges(p sse.Publisher) {
	events.AppCreated.Register(newBridge(p, fixedChannel[events.AppCreatedPayload]("apps"), payloadContent))
	events.TestCreated.Register(newBridge(p, fixedChannel[events.TestCreatedPayload]("tests"), payloadContent))
	events.NewTestLogEntry.Register(newBridge(p, idChannel("test_run_%d_log", func(e events.NewTestLogEntryPayload) uint { return e.TestRunID }), payloadContent))
}

func registerProtocolBridges(p sse.Publisher) {
	events.NewTestProtocol.Register(newBridge(p, idChannel("test_run_%d_protocol", func(e events.NewTestProtocolPayload) uint { return e.TestRunID }), protocolContent))
	events.NewTestProtocolLog.Register(newBridge(p, idChannel("test_protocol_%d_log", func(e events.NewTestProtocolLogPayload) uint { return e.TestProtocolID }), payloadContent))
	events.ProtocolKey.Register(newBridge(p, idChannel("test_protocol_%d_keys", func(e events.ProtocolKeyPayload) uint { return e.TestProtocolID }), payloadContent))
}

func registerTestRunBridges(p sse.Publisher) {
	events.TestRunFinished.Register(newBridge(p, idChannel("test_run_%d_finished", func(e events.TestRunFinishedPayload) uint { return e.TestRunID }), payloadContent))
	registerTestRunChangeBridges(p)
}

func registerTestRunChangeBridges(p sse.Publisher) {
	const channel = "test_runs"
	events.TestRunStarted.Register(newBridge(p, fixedChannel[events.TestRunStartedPayload](channel), startedChange))
	events.TestRunFinished.Register(newBridge(p, fixedChannel[events.TestRunFinishedPayload](channel), finishedChange))
}

func registerNodeBridges(p sse.Publisher) {
	events.NodeConnected.Register(newBridge(p, fixedChannel[events.NodeConnectedPayload]("nodes"), nodeConnectedStatus))
	events.NodeDisconnected.Register(newBridge(p, fixedChannel[events.NodeDisconnectedPayload]("nodes"), nodeDisconnectedStatus))
}

func startedChange(e events.TestRunStartedPayload) any {
	return testRunChange{ProjectID: e.ProjectID, AppID: e.AppID, TestID: e.TestID, TestRunID: e.TestRunID, State: "started"}
}

func finishedChange(e events.TestRunFinishedPayload) any {
	return testRunChange{ProjectID: e.ProjectID, AppID: e.AppID, TestRunID: e.TestRunID, State: "finished"}
}

func nodeConnectedStatus(e events.NodeConnectedPayload) any {
	return NodeStatusPayload{NodeIdentifier: e.NodeIdentifier, Connected: true}
}

func nodeDisconnectedStatus(e events.NodeDisconnectedPayload) any {
	return NodeStatusPayload{NodeIdentifier: e.NodeIdentifier, Connected: false}
}

// protocolContent sends the protocol the way the UI can encode it.
func protocolContent(e events.NewTestProtocolPayload) any {
	if p, ok := e.Protocol.(*models.TestProtocol); ok {
		return protocolForUI(p)
	}
	return e.Protocol
}

// protocolForUI copies the protocol without the live device driver and
// connection: a driver points back to its device config, so the protocol
// can't be encoded as JSON and the run page would never see the test.
func protocolForUI(p *models.TestProtocol) *models.TestProtocol {
	c := *p
	if p.Device != nil {
		d := *p.Device
		d.Dev = nil
		d.Connection = nil
		c.Device = &d
	}
	return &c
}
