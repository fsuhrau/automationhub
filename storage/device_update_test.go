package storage

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

type fakeNotifier struct {
	lists    []events.DeviceListChangedPayload
	statuses []events.DeviceStatusChangedPayload
}

func (f *fakeNotifier) DeviceListChanged(p events.DeviceListChangedPayload) {
	f.lists = append(f.lists, p)
}
func (f *fakeNotifier) DeviceStatusChanged(p events.DeviceStatusChangedPayload) {
	f.statuses = append(f.statuses, p)
}

func newStoredDevice(t *testing.T) (*deviceStore, *fakeNotifier, *testkit.FakeDevice) {
	t.Helper()
	db := testkit.DB(t)
	n := &fakeNotifier{}
	store := NewDeviceStoreWithNotifier(db, n)
	dev := testkit.NewFakeDevice("dev-1")
	dev.OSName, dev.OSVersion = "android", "14"
	dev.Parameter = map[string]string{"abi": "arm64"}
	if err := store.NewDevice("m", models.Device{DeviceIdentifier: "dev-1", Name: "fake", OS: "android", OSVersion: "14", Manager: "m",
		DeviceParameter: []models.DeviceParameter{{Key: "abi", Value: "arm64"}}}); err != nil {
		t.Fatal(err)
	}
	n.lists = nil
	return store, n, dev
}

func TestUpdateBranches(t *testing.T) {
	tests := []struct {
		name         string
		change       func(d *testkit.FakeDevice)
		wantList     int
		wantStatus   int
		wantLogCount int64
	}{
		{"nothing changed", func(d *testkit.FakeDevice) {}, 0, 0, 0},
		{"name", func(d *testkit.FakeDevice) { d.Name = "new" }, 1, 0, 0},
		{"os", func(d *testkit.FakeDevice) { d.OSName = "ios" }, 1, 0, 0},
		{"os version", func(d *testkit.FakeDevice) { d.OSVersion = "15" }, 1, 0, 0},
		{"parameter value", func(d *testkit.FakeDevice) { d.Parameter["abi"] = "x86" }, 1, 0, 0},
		{"new parameter", func(d *testkit.FakeDevice) { d.Parameter["api"] = "34" }, 1, 0, 0},
		{"nil parameters", func(d *testkit.FakeDevice) { d.Parameter = nil }, 0, 0, 0},
		{"status only", func(d *testkit.FakeDevice) { d.State = device.StateBooted }, 0, 1, 1},
		{"status and name", func(d *testkit.FakeDevice) { d.State = device.StateBooted; d.Name = "n" }, 1, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, n, dev := newStoredDevice(t)
			tt.change(dev)
			if err := store.Update("m", dev); err != nil {
				t.Fatal(err)
			}
			if len(n.lists) != tt.wantList || len(n.statuses) != tt.wantStatus {
				t.Fatalf("lists=%d statuses=%d, want %d/%d", len(n.lists), len(n.statuses), tt.wantList, tt.wantStatus)
			}
			var logs int64
			store.db.Model(&models.DeviceLog{}).Count(&logs)
			if logs != tt.wantLogCount {
				t.Fatalf("device logs = %d, want %d", logs, tt.wantLogCount)
			}
		})
	}
}

func TestUpdateManagerChangeNotifiesAndPersists(t *testing.T) {
	store, n, dev := newStoredDevice(t)
	got, _ := store.GetDevice("m", "dev-1")
	got.Manager = "old"
	if err := store.Update("m", dev); err != nil {
		t.Fatal(err)
	}
	if len(n.lists) != 1 || got.Manager != "m" {
		t.Fatalf("lists=%d manager=%s", len(n.lists), got.Manager)
	}
}

func TestUpdateUnknownDevice(t *testing.T) {
	store, n, _ := newStoredDevice(t)
	if err := store.Update("m", testkit.NewFakeDevice("other")); err == nil {
		t.Fatal("expected not found error")
	}
	if len(n.lists)+len(n.statuses) != 0 {
		t.Fatal("notified for unknown device")
	}
}

func TestUpdateSaveFailureDoesNotNotify(t *testing.T) {
	store, n, dev := newStoredDevice(t)
	dev.Name = "new"
	if err := store.db.Migrator().DropTable(&models.Device{}); err != nil {
		t.Fatal(err)
	}
	if err := store.Update("m", dev); err == nil {
		t.Fatal("expected save error")
	}
	if len(n.lists) != 0 {
		t.Fatal("notified despite failed save")
	}
}

func TestNewDeviceNotifiesAdded(t *testing.T) {
	db := testkit.DB(t)
	n := &fakeNotifier{}
	store := NewDeviceStoreWithNotifier(db, n)
	if err := store.NewDevice("m", models.Device{DeviceIdentifier: "a"}); err != nil {
		t.Fatal(err)
	}
	if len(n.lists) != 1 || !n.lists[0].Added {
		t.Fatalf("lists = %+v", n.lists)
	}
}

// The default store announces on the event bus.
func TestDefaultStoreTriggersEventBus(t *testing.T) {
	lists := make(chan events.DeviceListChangedPayload, 4)
	statuses := make(chan events.DeviceStatusChangedPayload, 4)
	events.DeviceListChanged.Register(events.HandlerFunc[events.DeviceListChangedPayload](func(p events.DeviceListChangedPayload) { lists <- p }))
	events.DeviceStatusChanged.Register(events.HandlerFunc[events.DeviceStatusChangedPayload](func(p events.DeviceStatusChangedPayload) { statuses <- p }))
	store := NewDeviceStore(testkit.DB(t))
	if err := store.NewDevice("m", models.Device{DeviceIdentifier: "a", Name: "fake"}); err != nil {
		t.Fatal(err)
	}
	<-lists
	dev := testkit.NewFakeDevice("a")
	dev.State = device.StateBooted
	if err := store.Update("m", dev); err != nil {
		t.Fatal(err)
	}
	if got := <-statuses; got.DeviceState != uint(device.StateBooted) {
		t.Fatalf("status payload = %+v", got)
	}
}

func TestUpdateLogFailureDoesNotNotify(t *testing.T) {
	store, n, dev := newStoredDevice(t)
	dev.State = device.StateBooted
	if err := store.db.Migrator().DropTable(&models.DeviceLog{}); err != nil {
		t.Fatal(err)
	}
	if err := store.Update("m", dev); err == nil {
		t.Fatal("expected log error")
	}
	if len(n.statuses) != 0 {
		t.Fatal("notified despite failed log")
	}
}
