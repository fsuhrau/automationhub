package hub

import (
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// The optional per-device roles of a node connection (manager.NodeRecorder,
// NodeScreenshotter, NodeFeatures, NodeNativeScripter). A connection that
// doesn't offer the role makes the call fail with device.ErrUnsupported; the
// calls that have no error are dropped like those to an unreachable node.

var (
	_ manager.NodeRecorder       = (*NodeManager)(nil)
	_ manager.NodeScreenshotter  = (*NodeManager)(nil)
	_ manager.NodeFeatures       = (*NodeManager)(nil)
	_ manager.NodeNativeScripter = (*NodeManager)(nil)
)

// forwardRole is forward for calls that need the node's connection to offer
// the role R.
func forwardRole[R, T any](nm *NodeManager, id manager.NodeIdentifier, fn func(R) (T, error)) (T, error) {
	return forward(nm, id, func(h manager.RPCClient) (T, error) {
		role, ok := h.(R)
		if !ok {
			var zero T
			return zero, fmt.Errorf("node %s: %w", id, device.ErrUnsupported)
		}
		return fn(role)
	})
}

// forwardRoleErr is forwardRole for calls that only return an error.
func forwardRoleErr[R any](nm *NodeManager, id manager.NodeIdentifier, fn func(R) error) error {
	_, err := forwardRole(nm, id, func(r R) (struct{}, error) { return struct{}{}, fn(r) })
	return err
}

// forwardRoleOr is forwardRole for calls without an error: fallback is
// returned when the node can't be reached or lacks the role.
func forwardRoleOr[R, T any](at nodeRef, fallback T, fn func(R) T) T {
	value, err := forwardRole(at.nm, at.id, func(r R) (T, error) { return fn(r), nil })
	if err != nil {
		return fallback
	}
	return value
}

// forwardRoleDo is forwardRole for fire-and-forget calls.
func forwardRoleDo[R any](nm *NodeManager, id manager.NodeIdentifier, fn func(R)) {
	forwardRoleOr(nm.node(id), struct{}{}, func(r R) struct{} { fn(r); return struct{}{} })
}

// StartRecording ignores path: the node decides where a recording goes.
func (nm *NodeManager) StartRecording(id manager.NodeIdentifier, deviceId string, _ string) error {
	return forwardRoleErr(nm, id, func(r manager.RPCRecorder) error { return r.StartRecording(deviceId, "") })
}

func (nm *NodeManager) StopRecording(id manager.NodeIdentifier, deviceId string) error {
	return forwardRoleErr(nm, id, func(r manager.RPCRecorder) error { return r.StopRecording(deviceId) })
}

func (nm *NodeManager) GetScreenshot(id manager.NodeIdentifier, deviceId string) ([]byte, int, int, error) {
	shot, err := forwardRole(nm, id, func(r manager.RPCScreenshotter) (screenshot, error) {
		data, width, height, err := r.GetScreenshot(deviceId)
		return screenshot{data, width, height}, err
	})
	return shot.data, shot.width, shot.height, err
}

// screenshot bundles the results of GetScreenshot for forwardRole.
type screenshot struct {
	data          []byte
	width, height int
}

func (nm *NodeManager) HasFeature(id manager.NodeIdentifier, deviceId string, feature string) bool {
	return forwardRoleOr(nm.node(id), false, func(r manager.RPCFeatures) bool { return r.HasFeature(deviceId, feature) })
}

func (nm *NodeManager) Execute(id manager.NodeIdentifier, deviceId string, data string) {
	forwardRoleDo(nm, id, func(r manager.RPCFeatures) { r.Execute(deviceId, data) })
}

func (nm *NodeManager) RunNativeScript(id manager.NodeIdentifier, deviceId string, script []byte) {
	forwardRoleDo(nm, id, func(r manager.RPCNativeScripter) { r.RunNativeScript(deviceId, script) })
}
