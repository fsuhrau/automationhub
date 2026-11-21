package generic

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/device"
)

func TestInfoGetters(t *testing.T) {
	i := Info{
		ID: "id", Name: "name", OSName: "os", OSVersion: "1.2", State: device.StateBooted,
		Type: 3, Platform: 4, Parameter: map[string]string{"k": "v"}, Target: "9",
	}
	got := []any{
		i.DeviceID(), i.DeviceName(), i.DeviceOSName(), i.DeviceOSVersion(), i.DeviceState(),
		i.DeviceType(), i.PlatformType(), i.DeviceParameter(), i.TargetVersion(),
	}
	want := []any{"id", "name", "os", "1.2", device.StateBooted, 3, 4, map[string]string{"k": "v"}, "9"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getters = %v, want %v", got, want)
	}
}
