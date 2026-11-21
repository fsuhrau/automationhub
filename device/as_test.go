package device_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestAs(t *testing.T) {
	d := testkit.NewFakeDevice("a")

	if r, ok := device.As[device.Recorder](d); !ok || r == nil {
		t.Error("FakeDevice should be a Recorder")
	}
	if _, ok := device.As[device.AppDataClearer](d); !ok {
		t.Error("FakeDevice should be an AppDataClearer")
	}
	if _, ok := device.As[device.Handler](d); ok {
		t.Error("FakeDevice must not be a Handler")
	}
}

func TestDeviceIsTheCompositeOfTheRoles(t *testing.T) {
	var d device.Device = testkit.NewFakeDevice("a")
	var _ device.Identity = d
	var _ device.AppManager = d
	var _ device.AppConnection = d
	var _ device.Locker = d
	var _ device.TestLogger = d
	var _ device.Recorder = d
	var _ device.Screenshotter = d
	var _ device.FeatureExecutor = d
	var _ device.NativeScripter = d
	var _ device.ActionHandlerRegistry = d
}
