package unityeditor

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

func registerData(id string) device.RegisterData {
	return device.RegisterData{
		DeviceID: id, Name: "host", DeviceOS: "macOS", DeviceOSVersion: "14",
		TargetVersion: testVersion, DeviceOSInfos: "infos", ProjectDir: "/proj",
		DeviceIP: net.ParseIP("10.0.0.2"),
	}
}

func TestRegisterDeviceCreatesAndStores(t *testing.T) {
	env := newEnv(nil)
	got, err := env.handler.RegisterDevice(registerData("e1"))
	if err != nil {
		t.Fatal(err)
	}
	if got.DeviceName() != "Unity ("+testVersion+") @ host" || got.TargetVersion() != testVersion || got.DeviceOSName() != "macOS" {
		t.Errorf("device = %s %s %s", got.DeviceName(), got.TargetVersion(), got.DeviceOSName())
	}
	if !env.handler.HasDevice(got) {
		t.Fatal("the editor must be listed")
	}
	if len(env.storage.created) != 1 {
		t.Fatalf("created = %d", len(env.storage.created))
	}
	m := env.storage.created[0]
	if m.DeviceIdentifier != "e1" || m.DeviceType != models.DeviceTypeUnityEditor || m.Manager != Manager ||
		m.GetAttribute("projectDir") != "/proj" || m.ConnectionParameter.ConnectionType != models.ConnectionTypeRemote ||
		m.OSInfos != "infos" || m.OSVersion != "14" {
		t.Errorf("model = %+v", m)
	}
	if dev, _ := env.handler.Get("e1"); dev.GetConfig().GetAttribute("projectDir") != "/proj" {
		t.Error("the config carries the project dir")
	}
}

func TestRegisterDeviceUpdatesAKnownEditor(t *testing.T) {
	env := newEnv(nil)
	first, _ := env.handler.RegisterDevice(registerData("e1"))
	data := registerData("e1")
	data.Name, data.DeviceOSVersion = "other", "15"
	second, err := env.handler.RegisterDevice(data)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || second.DeviceName() != "Unity ("+testVersion+") @ other" || second.DeviceOSVersion() != "15" {
		t.Fatalf("second = %s %s", second.DeviceName(), second.DeviceOSVersion())
	}
	if len(env.storage.created) != 1 || env.storage.updateCount() != 1 {
		t.Fatalf("created = %d, updated = %d", len(env.storage.created), env.storage.updateCount())
	}
}

func TestRegisterDeviceKeepsGoingOnStorageErrors(t *testing.T) {
	env := newEnv(nil)
	env.storage.newErr, env.storage.updateErr = errBoom, errBoom
	for i := 0; i < 2; i++ {
		if _, err := env.handler.RegisterDevice(registerData("e1")); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
	}
}

func TestConnOfNilStaysNil(t *testing.T) {
	if connOf(nil) != nil {
		t.Fatal("a nil websocket must give a nil managerConn")
	}
}

func TestRefreshDevices(t *testing.T) {
	tests := []struct {
		name       string
		silent     time.Duration
		state      device.State
		updated    bool
		wantStored int
		wantState  device.State
	}{
		{"fresh and unchanged", time.Second, device.StateBooted, false, 0, device.StateBooted},
		{"fresh and updated", time.Second, device.StateBooted, true, 1, device.StateBooted},
		{"silent too long", 2 * time.Minute, device.StateBooted, false, 1, device.StateShutdown},
		{"silent but already shut down", 2 * time.Minute, device.StateShutdown, false, 0, device.StateShutdown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newEnv(nil)
			dev := env.addEditor("e1", "")
			dev.State, dev.updated = tt.state, tt.updated
			dev.lastUpdateAt = epochTime().Add(-tt.silent)
			env.storage.updateErr = errBoom

			if err := env.handler.RefreshDevices(false); err != nil {
				t.Fatal(err)
			}
			if env.storage.updateCount() != tt.wantStored || dev.DeviceState() != tt.wantState || dev.updated {
				t.Fatalf("stored = %d, state = %v, updated = %v", env.storage.updateCount(), dev.DeviceState(), dev.updated)
			}
		})
	}
}

func TestInit(t *testing.T) {
	env := newEnv(nil)
	env.storage.stored = models.Devices{&models.Device{
		DeviceIdentifier: "e1", Name: "Editor", OS: "macOS", OSVersion: "14", OSInfos: "i", TargetVersion: testVersion,
	}}
	if err := env.handler.Init("master", "node", nil); err != nil {
		t.Fatal(err)
	}
	dev, ok := env.handler.Get("e1")
	if !ok || dev.DeviceName() != "Editor" || dev.TargetVersion() != testVersion || dev.osInfos != "i" {
		t.Fatalf("device = %+v", dev)
	}
	if dev.GetLogWriter() == nil || len(dev.ActionHandlers()) != 1 || dev.GetConfig() == nil {
		t.Error("a loaded editor reports to the master")
	}
	if dev.DeviceState() != device.StateShutdown || env.storage.updateCount() != 1 {
		t.Errorf("a loaded editor that never reported is shut down and stored, state = %v", dev.DeviceState())
	}
}

func TestInitStorageError(t *testing.T) {
	env := newEnv(nil)
	env.storage.getErr = errBoom
	if err := env.handler.Init("m", "n", nil); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestHandlerBasics(t *testing.T) {
	h := NewHandler(config.Manager{}, &fakeStorage{})
	if h.Name() != Manager || h.Start() != nil || h.Stop() != nil {
		t.Fatal("handler basics")
	}
	if h.runner == nil || h.clock == nil || h.goos == "" {
		t.Fatal("defaults must be set")
	}
}
