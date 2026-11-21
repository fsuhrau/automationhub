package hostinfotest

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestStoreRecordsUpdates(t *testing.T) {
	boom := errors.New("boom")
	s := NewStore(boom)
	d := testkit.NewFakeDevice("a")
	if err := s.Update("m", d); !errors.Is(err, boom) {
		t.Fatalf("Update error %v, want boom", err)
	}
	if got := s.Updated(); len(got) != 1 || got[0] != device.Device(d) {
		t.Fatalf("Updated() = %v", got)
	}
}

func TestStoreReadsAreEmpty(t *testing.T) {
	s := NewStore(nil)
	if devs, err := s.GetDevices("m"); devs != nil || err != nil {
		t.Fatalf("GetDevices = %v, %v", devs, err)
	}
	if err := s.NewDevice("m", models.Device{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetDevice("m", "x"); !errors.Is(err, device.DeviceNotFoundError) {
		t.Fatalf("GetDevice error %v", err)
	}
}
