package base

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

type stateDevice struct {
	device.Device
	state device.State
}

func (d *stateDevice) DeviceState() device.State { return d.state }

// Only devices the node sees can run tests; network devices are no
// exception, they are connected with adb beforehand.
func TestUnavailable(t *testing.T) {
	for _, c := range []struct {
		name   string
		dev    models.Device
		reason string
	}{
		{"booted", models.Device{Dev: &stateDevice{state: device.StateBooted}}, ""},
		{"locked", models.Device{Dev: &stateDevice{state: device.StateLocked}}, ""},
		{"shut down simulator", models.Device{Dev: &stateDevice{state: device.StateShutdown}}, ""},
		{"unplugged", models.Device{Dev: &stateDevice{state: device.StateRemoteDisconnected}}, "disconnected"},
		{"network device not connected", models.Device{ConnectionType: models.ConnectionTypeRemote,
			Dev: &stateDevice{state: device.StateRemoteDisconnected}}, "disconnected"},
		{"node offline", models.Device{Dev: &stateDevice{state: device.StateNodeDisconnected}}, "on a node that is offline"},
		{"unauthorized", models.Device{Dev: &stateDevice{state: device.StateUnauthorized}}, "not authorized for USB debugging"},
		{"unknown to the hub", models.Device{}, "not connected to the hub"},
	} {
		if got := Unavailable(c.dev); got != c.reason {
			t.Errorf("%s: got %q, want %q", c.name, got, c.reason)
		}
	}
}
