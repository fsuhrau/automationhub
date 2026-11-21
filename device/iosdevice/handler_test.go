package iosdevice

import (
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

func TestRefreshAddsReadableAndLockedDevices(t *testing.T) {
	r := newRig(t, "A", "B")
	r.usb.locked = map[string]bool{"B": true}
	if err := r.h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	a, b := r.device(t, "A"), r.device(t, "B")
	if a.DeviceState() != device.StateBooted || a.DeviceName() != "iPhone of A" || a.DeviceOSName() != "iPhone OS" || a.DeviceOSVersion() != "18.1" {
		t.Errorf("A = %v %q %q %q", a.DeviceState(), a.DeviceName(), a.DeviceOSName(), a.DeviceOSVersion())
	}
	if b.DeviceState() != device.StateUnauthorized || b.DeviceName() != "B" || b.DeviceOSName() != "iOS" {
		t.Errorf("B = %v %q %q", b.DeviceState(), b.DeviceName(), b.DeviceOSName())
	}
	if len(r.store.created) != 2 || r.store.created[0].DeviceIdentifier != "A" || r.store.created[0].Name != "iPhone of A" ||
		r.store.created[0].PlatformType != models.PlatformTypeiOS || r.store.created[0].Manager != Manager {
		t.Errorf("created = %+v", r.store.created)
	}
	if a.GetConfig() == nil || len(a.ActionHandlers()) != 1 || a.GetLogWriter() == nil {
		t.Error("the device isn't wired to the master")
	}
	if got := r.store.lastUpdates(2); got[0] != "A booted" || got[1] != "B unauthorized" {
		t.Errorf("updates = %v", got)
	}
	devices, err := r.h.GetDevices()
	if err != nil || len(devices) != 2 || !r.h.HasDevice(a) {
		t.Errorf("GetDevices = %v, %v", devices, err)
	}
}

func TestRefreshUpdatesKnownDevicesAndMarksMissingOnesDisconnected(t *testing.T) {
	r := newRig(t, "A", "B")
	r.usb.locked = map[string]bool{"B": true}
	_ = r.h.RefreshDevices(true)
	r.clock.Advance(time.Second)

	r.usb.locked = nil // B is trusted now
	r.usb.serials = []string{"B"}
	if err := r.h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	if got := r.device(t, "B").DeviceState(); got != device.StateBooted {
		t.Errorf("B state = %v", got)
	}
	if got := r.device(t, "B").DeviceName(); got != "iPhone of B" {
		t.Errorf("B name = %q", got)
	}
	if got := r.device(t, "A").DeviceState(); got != device.StateRemoteDisconnected {
		t.Errorf("A state = %v", got)
	}
	if len(r.store.created) != 2 {
		t.Errorf("known devices were created again: %d", len(r.store.created))
	}
	if got := r.store.lastUpdates(2); got[0] != "B booted" || got[1] != "A disconnected" {
		t.Errorf("updates = %v", got)
	}
}

func TestRefreshKeepsKnownLockedDeviceNames(t *testing.T) {
	r := newRig(t, "A")
	_ = r.h.RefreshDevices(true)
	r.clock.Advance(time.Second)
	r.usb.locked = map[string]bool{"A": true}
	_ = r.h.RefreshDevices(true)
	a := r.device(t, "A")
	if a.DeviceState() != device.StateUnauthorized || a.DeviceName() != "iPhone of A" {
		t.Errorf("A = %v %q", a.DeviceState(), a.DeviceName())
	}
}

func TestRefreshReportsListErrors(t *testing.T) {
	r := newRig(t, "A")
	r.usb.listErr = errors.New("usbmuxd not running")
	if err := r.h.RefreshDevices(true); err == nil {
		t.Fatal("expected an error")
	}
	if _, ok := r.h.Get("A"); ok {
		t.Error("a device appeared although listing failed")
	}
}

func TestUnforcedRefreshIsASafetyNetWhileUsbmuxdIsTracked(t *testing.T) {
	r := newRig(t, "A")
	r.h.tracking.Store(true)
	_ = r.h.RefreshDevices(false) // first poll
	_ = r.h.RefreshDevices(false) // too soon
	r.clock.Advance(fallbackPollInterval)
	_ = r.h.RefreshDevices(false) // due again
	_ = r.h.RefreshDevices(true)  // forced
	if r.usb.lists != 3 {
		t.Errorf("devices were listed %d times, want 3", r.usb.lists)
	}
	r.h.tracking.Store(false)
	_ = r.h.RefreshDevices(false) // without events every poll reads
	if r.usb.lists != 4 {
		t.Errorf("devices were listed %d times, want 4", r.usb.lists)
	}
}

func TestInitLoadsStoredDevices(t *testing.T) {
	r := newRig(t)
	r.store.stored = models.Devices{{DeviceIdentifier: "S1"}, {DeviceIdentifier: "S2"}}
	token := "tok"
	if err := r.h.Init("master:1", "node", &token); err != nil {
		t.Fatal(err)
	}
	d := r.device(t, "S1")
	if d.GetConfig() != r.store.stored[0] || len(d.ActionHandlers()) != 1 || d.GetLogWriter() == nil {
		t.Error("stored device isn't set up")
	}
	if !r.h.Has("S2") || r.h.Name() != Manager {
		t.Error("S2 missing")
	}
}

func TestStartAndStopUnknownDevice(t *testing.T) {
	r := newRig(t)
	if err := r.h.StartDevice("nope"); !errors.Is(err, device.DeviceNotFoundError) {
		t.Errorf("StartDevice = %v", err)
	}
	if err := r.h.StopDevice("nope"); !errors.Is(err, device.DeviceNotFoundError) {
		t.Errorf("StopDevice = %v", err)
	}
}

func TestRegisterDeviceIsUnsupported(t *testing.T) {
	if _, err := newRig(t).h.RegisterDevice(device.RegisterData{}); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}

func TestInitReportsStorageErrors(t *testing.T) {
	r := newRig(t)
	r.h.deviceStorage = failingStore{r.store}
	if err := r.h.Init("m", "n", nil); err == nil {
		t.Fatal("expected an error")
	}
}

type failingStore struct{ *fakeStore }

func (failingStore) GetDevices(string) (models.Devices, error) { return nil, errors.New("db down") }

func TestStartWatchesUsbmuxdAndRefreshesAgainWhenTheDeviceSettled(t *testing.T) {
	r := newRig(t, "A")
	changes := make(chan struct{}, 10)
	r.h.SetChangeNotifier(func() { changes <- struct{}{} })
	r.h.deps.openListener = func() (muxListener, error) { return newBlockingListener(), nil }
	if err := r.h.Start(); err != nil {
		t.Fatal(err)
	}
	<-changes // the refresh after the connect
	waitFor(t, func() bool { return r.h.Has("A") && r.h.tracking.Load() })
	waitPending(t, r.clock, 1)
	r.clock.Advance(settleDelay)
	<-changes // the refresh once the device settled
	if err := r.h.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := r.h.Stop(); err != nil { // stopping twice is fine
		t.Fatal(err)
	}
}

func TestStopEndsThePendingSettleRefresh(t *testing.T) {
	r := newRig(t, "A")
	changes := make(chan struct{}, 10)
	r.h.SetChangeNotifier(func() { changes <- struct{}{} })
	r.h.deps.openListener = func() (muxListener, error) { return newBlockingListener(), nil }
	_ = r.h.Start()
	<-changes
	waitPending(t, r.clock, 1)
	_ = r.h.Stop()
	r.clock.Advance(settleDelay)
	select {
	case <-changes:
		t.Fatal("refreshed after Stop")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRefreshAndNotifyNotifiesEvenWhenTheRefreshFails(t *testing.T) {
	r := newRig(t)
	r.usb.listErr = errors.New("down")
	notified := 0
	r.h.SetChangeNotifier(func() { notified++ })
	r.h.refreshAndNotify()
	if notified != 1 {
		t.Errorf("notified %d times", notified)
	}
	r.h.notifyChange = nil
	r.h.changed() // no notifier: nothing happens
}

func newBlockingListener() *fakeListener {
	return &fakeListener{events: make(chan bool), closed: make(chan struct{})}
}

func TestNewHandlerAddsTheHookToTheStandardLogger(t *testing.T) {
	h := NewHandler(config.Manager{}, &fakeStore{})
	got, err := h.Hook().Await(time.Minute, func(time.Duration) <-chan time.Time { return nil }, func(func(error)) {
		logrus.StandardLogger().SetOutput(discardWriter{})
		logrus.Info("ServerURLHere->http://h:9<-ServerURLHere")
	})
	if err != nil || got != "http://h:9" {
		t.Fatalf("Await = %q, %v", got, err)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestDepsDefaults(t *testing.T) {
	d := Deps{}.withDefaults()
	if d.Clock == nil || d.HTTP == nil || d.Processes == nil || d.Hook == nil || d.openListener == nil {
		t.Errorf("deps = %+v", d)
	}
	u := d.USB
	if u.ListDevices == nil || u.GetValues == nil || u.GetDevice == nil || u.Probe == nil || u.Forward == nil || u.RunXCUI == nil || u.CloseXCUI == nil {
		t.Errorf("usb = %+v", u)
	}
	custom := Deps{Hook: d.Hook}.withDefaults()
	if custom.Hook != d.Hook {
		t.Error("a given hook was replaced")
	}
}
