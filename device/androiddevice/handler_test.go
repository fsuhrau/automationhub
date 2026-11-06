package androiddevice

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// memoryStore is an in-memory storage.Device.
type memoryStore struct {
	devices models.Devices
}

func (s *memoryStore) GetDevices(string) (models.Devices, error) { return s.devices, nil }
func (s *memoryStore) NewDevice(_ string, d models.Device) error {
	s.devices = append(s.devices, &d)
	return nil
}
func (s *memoryStore) GetDevice(_ string, id string) (*models.Device, error) {
	for _, d := range s.devices {
		if d.DeviceIdentifier == id {
			return d, nil
		}
	}
	return nil, device.DeviceNotFoundError
}
func (s *memoryStore) Update(string, device.Device) error { return nil }

func TestApplyDeviceListFollowsAuthorization(t *testing.T) {
	h := NewHandler(&memoryStore{})
	state := func(serial string) device.State {
		t.Helper()
		d, ok := h.devices[serial]
		if !ok {
			t.Fatalf("device %s not known", serial)
		}
		return d.DeviceState()
	}

	h.applyDeviceList(parseADBDeviceList("RZCTA1YRJ5W            unauthorized usb:0-1.3 transport_id:2\n"))
	if s := state("RZCTA1YRJ5W"); s != device.StateUnauthorized {
		t.Fatalf("expected unauthorized, got %v", s)
	}
	if h.devices["RZCTA1YRJ5W"].deviceParameter != nil {
		t.Fatal("properties must not be read before authorization")
	}

	h.applyDeviceList(parseADBDeviceList("RZCTA1YRJ5W            device usb:0-1.3 product:r9q2xeea model:SM_G990B2 device:r9q transport_id:3\n"))
	d := h.devices["RZCTA1YRJ5W"]
	if d.DeviceState() != device.StateBooted || d.DeviceID() != "RZCTA1YRJ5W" || d.deviceParameter["Device Model"] != "SM_G990B2" {
		t.Fatalf("expected booted device with infos, got state=%v id=%q params=%v", d.DeviceState(), d.DeviceID(), d.deviceParameter)
	}

	h.applyDeviceList(nil) // unplugged
	if s := state("RZCTA1YRJ5W"); s != device.StateRemoteDisconnected {
		t.Fatalf("expected disconnected after unplug, got %v", s)
	}
}

func TestApplyDeviceListSetsIDOfStoredDevices(t *testing.T) {
	// devices restored from storage at startup have no id yet
	h := NewHandler(&memoryStore{devices: models.Devices{{DeviceIdentifier: "emulator-5554", Manager: Manager}}})
	h.devices["emulator-5554"] = &Device{deviceOSName: "android", installedApps: map[string]string{}}

	h.applyDeviceList(parseADBDeviceList("emulator-5554\tdevice product:sdk model:Pixel device:emu transport_id:1\n"))
	d := h.devices["emulator-5554"]
	if d.DeviceID() != "emulator-5554" || d.GetConfig() == nil || d.DeviceState() != device.StateBooted {
		t.Fatalf("stored device not restored: id=%q config=%v state=%v", d.DeviceID(), d.GetConfig(), d.DeviceState())
	}
}
