package api

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

type stateDevice struct {
	device.Device
	state device.State
}

func (d *stateDevice) DeviceState() device.State { return d.state }

// A run starts when at least one device can run it; without any it is
// refused with 409, so the caller (CI, CLI, MCP) fails at once.
func TestCheckDevices(t *testing.T) {
	booted := models.Device{Name: "Pixel", Dev: &stateDevice{state: device.StateBooted}}
	unplugged := models.Device{Name: "Samsung Galaxy S21", DeviceIdentifier: "RZCTA1YRJ5W", Dev: &stateDevice{state: device.StateRemoteDisconnected}}
	unknown := models.Device{DeviceIdentifier: "emulator-5554"}

	if err := checkDevices([]models.Device{unplugged, booted}); err != nil {
		t.Fatalf("one usable device: %v", err)
	}

	err := checkDevices([]models.Device{unplugged, unknown})
	if !errors.Is(err, ErrNoDevices) || statusForError(err) != http.StatusConflict {
		t.Fatalf("expected ErrNoDevices with 409, got %v", err)
	}
	for _, want := range []string{"Samsung Galaxy S21 is disconnected", "emulator-5554 is not connected to the hub"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q missing in %q", want, err)
		}
	}

	if err := checkDevices(nil); !errors.Is(err, ErrNoDevices) {
		t.Fatalf("expected ErrNoDevices without devices, got %v", err)
	}
}
