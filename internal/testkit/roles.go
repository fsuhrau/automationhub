package testkit

import "github.com/fsuhrau/automationhub/device"

// Roles lists the optional roles d implements, in a fixed order: Recorder,
// Screenshotter, FeatureExecutor, NativeScripter, AppDataClearer,
// InputCapable, LogReader, LogFollower, VideoStreamer. Platform tests compare
// it with the roles the platform really supports.
func Roles(d device.Device) []string {
	checks := []struct {
		name string
		has  func(device.Device) bool
	}{
		{"Recorder", hasRole[device.Recorder]},
		{"Screenshotter", hasRole[device.Screenshotter]},
		{"FeatureExecutor", hasRole[device.FeatureExecutor]},
		{"NativeScripter", hasRole[device.NativeScripter]},
		{"AppDataClearer", hasRole[device.AppDataClearer]},
		{"InputCapable", hasRole[device.InputCapable]},
		{"LogReader", hasRole[device.LogReader]},
		{"LogFollower", hasRole[device.LogFollower]},
		{"VideoStreamer", hasRole[device.VideoStreamer]},
	}
	var roles []string
	for _, c := range checks {
		if c.has(d) {
			roles = append(roles, c.name)
		}
	}
	return roles
}

func hasRole[R any](d device.Device) bool {
	_, ok := device.As[R](d)
	return ok
}
