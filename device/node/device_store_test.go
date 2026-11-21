package node

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestMemoryDeviceStore(t *testing.T) {
	store := NewMemoryDeviceStore(map[string]config.Manager{
		"android": {Devices: []config.Device{{Identifier: "a1", Parameter: map[string]string{"k": "v"}}}},
	})
	devs, err := store.GetDevices("android")
	if err != nil || len(devs) != 1 || devs[0].DeviceIdentifier != "a1" || devs[0].Manager != "android" || len(devs[0].CustomParameter) != 1 {
		t.Fatalf("GetDevices = %+v, %v", devs, err)
	}
	if _, err := store.GetDevices("ios"); !errors.Is(err, ErrManagerNotFound) {
		t.Errorf("unknown manager: %v", err)
	}
	if dev, err := store.GetDevice("android", "a1"); err != nil || dev.DeviceIdentifier != "a1" {
		t.Errorf("GetDevice = %+v, %v", dev, err)
	}
	if _, err := store.GetDevice("android", "zz"); !errors.Is(err, ErrDeviceNotFound) {
		t.Errorf("unknown device: %v", err)
	}
	if _, err := store.GetDevice("ios", "a1"); !errors.Is(err, ErrManagerNotFound) {
		t.Errorf("unknown manager: %v", err)
	}

	// NewDevice replaces a known device and adds a new one, also for a new manager
	_ = store.NewDevice("android", models.Device{DeviceIdentifier: "a1", Name: "replaced"})
	_ = store.NewDevice("android", models.Device{DeviceIdentifier: "a2"})
	_ = store.NewDevice("ios", models.Device{DeviceIdentifier: "i1"})
	if dev, _ := store.GetDevice("android", "a1"); dev.Name != "replaced" {
		t.Errorf("device not replaced: %+v", dev)
	}
	if devs, _ := store.GetDevices("android"); len(devs) != 2 {
		t.Errorf("android devices = %d", len(devs))
	}
	if _, err := store.GetDevice("ios", "i1"); err != nil {
		t.Error(err)
	}
	if err := store.Update("android", nil); err != nil {
		t.Error(err)
	}
}
