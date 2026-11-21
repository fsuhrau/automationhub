package device_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// assertsAs reports whether d is a role R, for the table below.
func assertsAs[R any](d device.Device) bool {
	_, ok := device.As[R](d)
	return ok
}

func TestAsFindsEveryOptionalRoleOfAFullDevice(t *testing.T) {
	d := testkit.NewFakeDevice("a")
	roles := map[string]bool{
		"Recorder":        assertsAs[device.Recorder](d),
		"Screenshotter":   assertsAs[device.Screenshotter](d),
		"FeatureExecutor": assertsAs[device.FeatureExecutor](d),
		"NativeScripter":  assertsAs[device.NativeScripter](d),
		"AppDataClearer":  assertsAs[device.AppDataClearer](d),
		"InputCapable":    assertsAs[device.InputCapable](d),
		"LogReader":       assertsAs[device.LogReader](d),
		"LogFollower":     assertsAs[device.LogFollower](d),
		"VideoStreamer":   assertsAs[device.VideoStreamer](d),
	}
	for name, ok := range roles {
		if !ok {
			t.Errorf("a full FakeDevice should be a %s", name)
		}
	}
	if assertsAs[device.Handler](d) {
		t.Error("FakeDevice must not be a Handler")
	}
}

func TestAsMissesTheOptionalRolesOfACoreDevice(t *testing.T) {
	d := testkit.NewCoreDevice("a")
	roles := map[string]bool{
		"Recorder":        assertsAs[device.Recorder](d),
		"Screenshotter":   assertsAs[device.Screenshotter](d),
		"FeatureExecutor": assertsAs[device.FeatureExecutor](d),
		"NativeScripter":  assertsAs[device.NativeScripter](d),
		"AppDataClearer":  assertsAs[device.AppDataClearer](d),
		"InputCapable":    assertsAs[device.InputCapable](d),
		"VideoStreamer":   assertsAs[device.VideoStreamer](d),
	}
	for name, ok := range roles {
		if ok {
			t.Errorf("a core device must not be a %s", name)
		}
	}
}

func TestDeviceIsTheCompositeOfTheCoreRoles(t *testing.T) {
	var d device.Device = testkit.NewCoreDevice("a")
	var _ device.Identity = d
	var _ device.AppManager = d
	var _ device.AppConnection = d
	var _ device.Locker = d
	var _ device.TestLogger = d
	var _ device.ActionHandlerRegistry = d
}
