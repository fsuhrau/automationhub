package testkit

import "github.com/fsuhrau/automationhub/device"

// CoreDevice is a device with the core roles only: Identity, AppManager,
// AppConnection, Locker, TestLogger and ActionHandlerRegistry. Unlike a
// FakeDevice it is none of the optional roles (Recorder, Screenshotter,
// FeatureExecutor, NativeScripter, InputCapable, ...), so device.As fails for
// them. It is for testing what a consumer does with a device that lacks a
// role. Fake is the device behind it, for scripting and for its recorded calls.
type CoreDevice struct {
	// Device is the core view of Fake. Embedding the interface hides the
	// optional methods of the FakeDevice.
	device.Device
	Fake *FakeDevice
}

// NewCoreDevice returns a core device with the given id.
func NewCoreDevice(id string) *CoreDevice {
	fake := NewFakeDevice(id)
	return &CoreDevice{Device: fake, Fake: fake}
}
