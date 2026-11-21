package iossim

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Functional: refresh with fake simctl output, then act on the simulator.
func TestRefreshThenAppActions(t *testing.T) {
	store := &memoryStore{}
	h, r, clock := newHandler(store)
	r.Script(listLine, []byte(listJSON), nil)
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}

	devs, err := h.GetDevices()
	if err != nil || len(devs) != 3 {
		t.Fatalf("devices: %d %v", len(devs), err)
	}
	d, _ := h.Get(udid1)
	if d.DeviceName() != "iPhone 15" || d.DeviceState() != device.StateBooted || d.DeviceOSName() != "iphonesimulator" || d.DeviceOSVersion() != "17.0" {
		t.Fatalf("unexpected simulator %+v", d.Info)
	}
	if !d.LastUpdate().Equal(clock.Now()) || !h.HasDevice(d) {
		t.Fatal("last update or registry wrong")
	}
	if len(store.devices) != 3 || store.devices[0].Manager != Manager || store.devices[0].PlatformType != models.PlatformTypeiOSSimulator {
		t.Fatalf("stored: %+v", store.devices)
	}

	p := &app.Parameter{Identifier: "com.example.game", App: &app.AppParams{AppPath: "/b/G.app"}}
	r.Script("xcrun simctl install AAAA-1111 /b/G.app", nil, nil)
	r.Script("xcrun simctl launch AAAA-1111 com.example.game SESSION_ID s NODE_URL n", nil, nil)
	r.Script("xcrun simctl terminate AAAA-1111 com.example.game", nil, nil)
	for _, err := range []error{d.InstallApp(p), d.StartApp(nil, p, "s", "n"), d.StopApp(p)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := r.Commands(); len(got) != 4 || got[3] != "xcrun simctl terminate AAAA-1111 com.example.game" {
		t.Fatalf("commands: %v", got)
	}
}

func TestRefreshUpdatesKnownAndMarksMissing(t *testing.T) {
	store := &memoryStore{}
	h, r, _ := newHandler(store)
	r.Script(listLine, []byte(listJSON), nil)
	if err := h.RefreshDevices(false); err != nil {
		t.Fatal(err)
	}
	r.Script(listLine, []byte(`{"devices":{"com.apple.CoreSimulator.SimRuntime.iOS-17-0":[{"state":"Shutdown","name":"iPhone 15 renamed","udid":"AAAA-1111"}]}}`), nil)
	if err := h.RefreshDevices(false); err != nil {
		t.Fatal(err)
	}
	a, _ := h.Get(udid1)
	if a.DeviceName() != "iPhone 15 renamed" || a.DeviceState() != device.StateShutdown {
		t.Fatalf("known simulator not updated: %+v", a.Info)
	}
	b, _ := h.Get(udid2)
	if b.DeviceState() != device.StateUnknown {
		t.Fatalf("missing local simulator must be unknown, got %v", b.DeviceState())
	}
	if len(store.devices) != 3 {
		t.Fatalf("known simulators must not be announced again: %d", len(store.devices))
	}
}

func TestMissingStateByConnectionType(t *testing.T) {
	remote := &models.Device{ConnectionParameter: &models.ConnectionParameter{ConnectionType: models.ConnectionTypeRemote}}
	tests := []struct {
		name string
		cfg  *models.Device
		want device.State
	}{
		{"no config", nil, device.StateUnknown},
		{"no connection parameter", &models.Device{}, device.StateUnknown},
		{"usb", &models.Device{ConnectionParameter: &models.ConnectionParameter{ConnectionType: models.ConnectionTypeUSB}}, device.StateUnknown},
		{"remote", remote, device.StateRemoteDisconnected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDevice(nil, false)
			d.SetConfig(tt.cfg)
			if got := missingState(d); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefreshErrors(t *testing.T) {
	h, r, _ := newHandler(&memoryStore{})
	r.Script(listLine, nil, errBoom)
	if err := h.RefreshDevices(false); !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
}

func TestRefreshSurvivesStorageErrors(t *testing.T) {
	store := &memoryStore{newErr: errBoom, updErr: errBoom}
	h, r, _ := newHandler(store)
	r.Script(listLine, []byte(listJSON), nil)
	if err := h.RefreshDevices(false); err != nil {
		t.Fatal(err)
	}
	if devs, _ := h.GetDevices(); len(devs) != 3 {
		t.Fatalf("devices: %d", len(devs))
	}
}
