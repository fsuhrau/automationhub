package unityeditor

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
)

func TestSetDeviceStateUsesParseState(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  device.State
	}{
		{"booted", "StateBooted", device.StateBooted},
		{"shutdown", "StateShutdown", device.StateShutdown},
		{"remote disconnected", "StateRemoteDisconnected", device.StateRemoteDisconnected},
		{"unknown name is StateUnknown, not shutdown", "bogus", device.StateUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{}
			d.SetDeviceState(tt.input)
			if got := d.DeviceState(); got != tt.want {
				t.Fatalf("state = %v, want %v", got, tt.want)
			}
		})
	}
}
