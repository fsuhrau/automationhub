package node

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func TestRPCDeviceIsOnlyADescription(t *testing.T) {
	d := &RPCDevice{Dev: &DeviceResponse{DeviceID: "d1"}}
	p := &app.Parameter{}
	installed, err := d.IsAppInstalled(p)
	calls := map[string]error{
		"IsAppInstalled": err,
		"InstallApp":     d.InstallApp(p),
		"UninstallApp":   d.UninstallApp(p),
		"StartApp":       d.StartApp(nil, p, "s", "u"),
		"StopApp":        d.StopApp(p),
	}
	for name, err := range calls {
		if !errors.Is(err, device.ErrUnsupported) {
			t.Errorf("%s: err = %v, want one wrapping device.ErrUnsupported", name, err)
		}
	}
	if installed || d.IsAppConnected() || d.ConnectionTimeout() == 0 {
		t.Error("a description is neither installed nor connected")
	}
}

func TestRPCDeviceHasNoOptionalRoles(t *testing.T) {
	var d device.Device = &RPCDevice{Dev: &DeviceResponse{}}
	if _, ok := device.As[device.Screenshotter](d); ok {
		t.Error("RPCDevice must not be a Screenshotter")
	}
	if _, ok := device.As[device.FeatureExecutor](d); ok {
		t.Error("RPCDevice must not be a FeatureExecutor")
	}
	if _, ok := device.As[device.Recorder](d); ok {
		t.Error("RPCDevice must not be a Recorder")
	}
}
