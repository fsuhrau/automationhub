package manager_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// The composite and every role are satisfied by the shared fake, so a consumer
// can depend on just the role it calls.
func TestDevicesRolesAreSatisfiedByTheFake(*testing.T) {
	fake := &testkit.FakeDeviceManager{}
	var (
		_ manager.Devices        = fake
		_ manager.DeviceRunner   = fake
		_ manager.DeviceLookup   = fake
		_ manager.DevicePower    = fake
		_ manager.DeviceRegistry = fake
		_ manager.ActionSender   = fake
	)
}
