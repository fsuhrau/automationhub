package node

import (
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
)

// The optional roles of a NodeDevice. The node decides what the real device
// can do, so the master side offers all of them and forwards. The node
// manager may not offer a role (manager.NodeRecorder and the like); the calls
// then fail with device.ErrUnsupported, or are skipped when they have no
// error to return.

var (
	_ device.Recorder        = (*NodeDevice)(nil)
	_ device.Screenshotter   = (*NodeDevice)(nil)
	_ device.FeatureExecutor = (*NodeDevice)(nil)
	_ device.NativeScripter  = (*NodeDevice)(nil)
)

// nodeRole is the node manager as role R, or an error wrapping
// device.ErrUnsupported.
func nodeRole[R any](d *NodeDevice) (R, error) {
	role, ok := d.nodeManager.(R)
	if !ok {
		return role, fmt.Errorf("node %s: %w", d.nodeId, device.ErrUnsupported)
	}
	return role, nil
}

// skipUnsupported logs a call that is dropped because the node manager lacks
// the role.
func (d *NodeDevice) skipUnsupported(call string, err error) {
	logrus.Debugf("%s of device %s skipped: %v", call, d.ID, err)
}

func (d *NodeDevice) StartRecording(path string) error {
	r, err := nodeRole[manager.NodeRecorder](d)
	if err != nil {
		return err
	}
	return r.StartRecording(d.nodeId, d.ID, path)
}

func (d *NodeDevice) StopRecording() error {
	r, err := nodeRole[manager.NodeRecorder](d)
	if err != nil {
		return err
	}
	return r.StopRecording(d.nodeId, d.ID)
}

func (d *NodeDevice) GetScreenshot() ([]byte, int, int, error) {
	r, err := nodeRole[manager.NodeScreenshotter](d)
	if err != nil {
		return nil, 0, 0, err
	}
	return r.GetScreenshot(d.nodeId, d.ID)
}

// HasFeature is false when the node manager has no features.
func (d *NodeDevice) HasFeature(feature string) bool {
	r, err := nodeRole[manager.NodeFeatures](d)
	if err != nil {
		d.skipUnsupported("HasFeature", err)
		return false
	}
	return r.HasFeature(d.nodeId, d.ID, feature)
}

// Execute is skipped when the node manager has no features.
func (d *NodeDevice) Execute(data string) {
	r, err := nodeRole[manager.NodeFeatures](d)
	if err != nil {
		d.skipUnsupported("Execute", err)
		return
	}
	r.Execute(d.nodeId, d.ID, data)
}

// RunNativeScript is skipped when the node manager has no native scripts.
func (d *NodeDevice) RunNativeScript(script []byte) {
	r, err := nodeRole[manager.NodeNativeScripter](d)
	if err != nil {
		d.skipUnsupported("RunNativeScript", err)
		return
	}
	r.RunNativeScript(d.nodeId, d.ID, script)
}
