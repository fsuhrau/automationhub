package hub

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// The forwarding methods of NodeManager all do the same: look up the RPC
// connection of the node and call it. forward, forwardErr, forwardOr and
// forwardDo are that one step, so each method is a single line.

// forward calls fn on the node's connection; without one it returns the
// lookup error (ErrNodeNotFound, ErrNodeNotConnected) and the zero value.
func forward[T any](nm *NodeManager, id manager.NodeIdentifier, fn func(manager.RPCClient) (T, error)) (T, error) {
	handler, err := nm.getHandler(id)
	if err != nil {
		var zero T
		return zero, err
	}
	return fn(handler)
}

// forwardErr is forward for calls that only return an error.
func forwardErr(nm *NodeManager, id manager.NodeIdentifier, fn func(manager.RPCClient) error) error {
	_, err := forward(nm, id, func(h manager.RPCClient) (struct{}, error) { return struct{}{}, fn(h) })
	return err
}

// forwardOr is forward for calls without an error: fallback is returned when
// the node can't be reached.
func forwardOr[T any](nm *NodeManager, id manager.NodeIdentifier, fallback T, fn func(manager.RPCClient) T) T {
	handler, err := nm.getHandler(id)
	if err != nil {
		return fallback
	}
	return fn(handler)
}

// forwardDo is forward for fire-and-forget calls; they are dropped when the
// node can't be reached.
func forwardDo(nm *NodeManager, id manager.NodeIdentifier, fn func(manager.RPCClient)) {
	forwardOr(nm, id, struct{}{}, func(h manager.RPCClient) struct{} { fn(h); return struct{}{} })
}

func (nm *NodeManager) StartDevice(id manager.NodeIdentifier, deviceId string) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.StartDevice(deviceId) })
}

func (nm *NodeManager) StopDevice(id manager.NodeIdentifier, deviceId string) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.StopDevice(deviceId) })
}

func (nm *NodeManager) IsAppInstalled(id manager.NodeIdentifier, deviceId string, p *app.Parameter) (bool, error) {
	return forward(nm, id, func(h manager.RPCClient) (bool, error) { return h.IsAppInstalled(deviceId, p) })
}

func (nm *NodeManager) IsAppUploaded(id manager.NodeIdentifier, p *app.Parameter) (bool, error) {
	return forward(nm, id, func(h manager.RPCClient) (bool, error) { return h.IsAppUploaded(p) })
}

// UploadApp uploads the app unless the node has it already.
func (nm *NodeManager) UploadApp(ctx context.Context, id manager.NodeIdentifier, p *app.Parameter) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error {
		exists, err := h.IsAppUploaded(p)
		if err != nil || exists {
			return err
		}
		return h.UploadApp(ctx, p)
	})
}

func (nm *NodeManager) InstallApp(id manager.NodeIdentifier, deviceId string, p *app.Parameter) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.InstallApp(deviceId, p) })
}

func (nm *NodeManager) UninstallApp(id manager.NodeIdentifier, deviceId string, p *app.Parameter) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.UninstallApp(deviceId, p) })
}

//nolint:revive // argument-limit: fixed by manager.NodeApps
func (nm *NodeManager) StartApp(id manager.NodeIdentifier, deviceId string, cfg *device.DeviceConfig, p *app.Parameter, sessionId string, nodeUrl string) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.StartApp(deviceId, cfg, p, sessionId, nodeUrl) })
}

func (nm *NodeManager) StopApp(id manager.NodeIdentifier, deviceId string, p *app.Parameter) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.StopApp(deviceId, p) })
}

func (nm *NodeManager) ClearAppData(id manager.NodeIdentifier, deviceId string, p *app.Parameter) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.ClearAppData(deviceId, p) })
}

func (nm *NodeManager) Control(id manager.NodeIdentifier, req device.ControlRequest) (device.ControlResponse, error) {
	return forward(nm, id, func(h manager.RPCClient) (device.ControlResponse, error) { return h.Control(req) })
}

func (nm *NodeManager) StartVideoStream(id manager.NodeIdentifier, req device.VideoStreamRequest) error {
	return forwardErr(nm, id, func(h manager.RPCClient) error { return h.StartVideoStream(req) })
}

func (nm *NodeManager) IsConnected(id manager.NodeIdentifier, deviceId string) bool {
	return forwardOr(nm, id, false, func(h manager.RPCClient) bool { return h.IsConnected(deviceId) })
}

func (nm *NodeManager) ConnectionTimeout(id manager.NodeIdentifier, deviceId string) time.Duration {
	return forwardOr(nm, id, NodeConnectionTimeout, func(h manager.RPCClient) time.Duration { return h.ConnectionTimeout(deviceId) })
}

func (nm *NodeManager) SendAction(id manager.NodeIdentifier, deviceId string, action []byte) {
	forwardDo(nm, id, func(h manager.RPCClient) { h.SendAction(deviceId, action) })
}
