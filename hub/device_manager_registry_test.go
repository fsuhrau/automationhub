package hub

import (
	"errors"
	"sort"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func registryFixture(t *testing.T) (*DeviceManager, *fakeHandler, *fakeHandler, *countingNotifier) {
	t.Helper()
	phone := testkit.NewFakeDevice("phone")
	phone.OSName = "android"
	sim := testkit.NewFakeDevice("sim")
	sim.OSName = "ios"
	android := newFakeHandler("android_device", phone)
	ios := newFakeHandler("ios_sim", sim)
	n := &countingNotifier{}
	dm := newTestManager(t)
	dm.notifier = n
	dm.AddHandler(android)
	dm.AddHandler(ios)
	return dm, android, ios, n
}

func TestRegistryAddAndListManagers(t *testing.T) {
	dm, _, _, _ := registryFixture(t)
	got := dm.Managers()
	sort.Strings(got)
	if len(got) != 2 || got[0] != "android_device" || got[1] != "ios_sim" {
		t.Fatalf("managers = %v", got)
	}
}

func TestRegistryDevicesGroupsByManager(t *testing.T) {
	dm, android, _, _ := registryFixture(t)
	android.listErr = errors.New("adb down")
	devices, err := dm.Devices()
	if err != nil || len(devices["ios_sim"]) != 1 || len(devices["android_device"]) != 1 {
		t.Fatalf("devices = %v, %v", devices, err)
	}
	dm.ListDevices() // logs both managers, including the failing one
}

func TestRegistryFindsDevices(t *testing.T) {
	dm, _, _, _ := registryFixture(t)
	tests := []struct {
		name   string
		lookup func() device.Device
		want   string
	}{
		{"get by id", func() device.Device { d, _ := dm.GetDevice("sim"); return d }, "sim"},
		{"unknown id", func() device.Device { d, _ := dm.GetDevice("nope"); return d }, ""},
		{"find by platform", func() device.Device { return dm.FindDevice(&app.Parameter{Platform: "android"}) }, "phone"},
		{"no platform match", func() device.Device { return dm.FindDevice(&app.Parameter{Platform: "web"}) }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lookup()
			if tt.want == "" {
				if got != nil {
					t.Fatalf("got %v, want none", got)
				}
				return
			}
			if got == nil || got.DeviceID() != tt.want {
				t.Fatalf("got %v, want %s", got, tt.want)
			}
		})
	}
	if _, manager := dm.GetDevice("sim"); manager != "ios_sim" {
		t.Fatalf("manager = %q", manager)
	}
}

func TestRegistryStartAndStopNotify(t *testing.T) {
	dm, android, _, n := registryFixture(t)
	phone, _ := dm.GetDevice("phone")
	if err := dm.Start(phone); err != nil {
		t.Fatal(err)
	}
	if err := dm.Stop(phone); err != nil {
		t.Fatal(err)
	}
	if len(android.started) != 1 || len(android.stopped) != 1 || n.notifications() != 2 {
		t.Fatalf("started %v stopped %v notified %d", android.started, android.stopped, n.notifications())
	}
}

func TestRegistryStartAndStopUnknownDevice(t *testing.T) {
	dm, _, _, n := registryFixture(t)
	ghost := testkit.NewFakeDevice("ghost")
	if err := dm.Start(ghost); !errors.Is(err, device.DeviceNotFoundError) {
		t.Fatalf("start err = %v", err)
	}
	if err := dm.Stop(ghost); !errors.Is(err, device.DeviceNotFoundError) {
		t.Fatalf("stop err = %v", err)
	}
	if n.notifications() != 0 {
		t.Fatal("nothing changed, nothing to notify")
	}
}

func TestRegistryRegisterDevicePicksManager(t *testing.T) {
	dm, android, _, _ := registryFixture(t)
	d, err := dm.RegisterDevice(device.RegisterData{ManagerType: "android_device", DeviceID: "new"})
	if err != nil || d.DeviceID() != "new" || android.registry.DeviceID != "new" {
		t.Fatalf("registered %v, %v", d, err)
	}
	if _, err := dm.RegisterDevice(device.RegisterData{ManagerType: "other"}); err == nil {
		t.Fatal("unknown manager type should fail")
	}
}
