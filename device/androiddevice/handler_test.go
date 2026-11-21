package androiddevice

import (
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

const (
	devicesLine = "adb devices -l"
	listBoth    = "List of devices attached\n" +
		"RZCTA1YRJ5W            device usb:0-1.3 product:r9q2xeea model:SM_G990B2 device:r9q transport_id:3\n" +
		"emulator-5554          device product:sdk model:Pixel device:emu transport_id:1\n" +
		"R58M1234               unauthorized usb:0-1.4 transport_id:5\n" +
		"192.168.0.9:5555       offline transport_id:6\n\n"
)

// scriptProps scripts the property reads of a device that is listed first.
func scriptProps(runner *testkit.FakeRunner, serial, release string) {
	p := "adb -s " + serial + " shell "
	runner.Script(p+"getprop ro.build.version.sdk", []byte("33\n"), nil)
	runner.Script(p+"getprop ro.build.version.release", []byte(release+"\n"), nil)
	runner.Script(p+"ip -f inet addr show wlan0", []byte(ipOutput), nil)
	runner.Script(p+"getprop ro.product.cpu.abilist", []byte("arm64-v8a\n"), nil)
}

func newTestHandler(store *memoryStore) (*Handler, *testkit.FakeRunner) {
	runner := &testkit.FakeRunner{}
	return NewHandlerWithDeps(store, testDeps(runner)), runner
}

func TestRefreshDevicesListsTheDevicesOfAdb(t *testing.T) {
	store := &memoryStore{}
	h, runner := newTestHandler(store)
	runner.Script(devicesLine, []byte(listBoth), nil)
	scriptProps(runner, "RZCTA1YRJ5W", "13")
	scriptProps(runner, "emulator-5554", "14")

	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}

	devs, _ := h.GetDevices()
	got := make([]string, len(devs))
	for i, d := range devs {
		got[i] = d.DeviceID() + "=" + device.StateToString(d.DeviceState())
	}
	// the offline device is not listed as a new device, the unauthorized one is
	if strings.Join(got, ",") != "R58M1234=unauthorized,RZCTA1YRJ5W=booted,emulator-5554=booted" {
		t.Fatalf("devices = %v", got)
	}
	phone, _ := h.Get("RZCTA1YRJ5W")
	if phone.DeviceName() != "RZCTA1YRJ5W" || phone.DeviceOSVersion() != "13" || phone.DeviceOSName() != "android" ||
		phone.DeviceParameter()["Device Model"] != "SM_G990B2" || phone.DeviceParameter()["API Level"] != "33" ||
		phone.DeviceType() != int(models.DeviceTypePhone) || phone.PlatformType() != int(models.PlatformTypeAndroid) {
		t.Fatalf("identity: %+v", phone.Info)
	}
	if len(store.created) != 3 || store.created[0].DeviceIdentifier != "RZCTA1YRJ5W" || store.created[0].OSVersion != "13" ||
		store.created[0].ConnectionParameter.ConnectionType != models.ConnectionTypeUSB {
		t.Fatalf("stored devices = %+v", store.created)
	}
	if phone.GetConfig() == nil || phone.GetLogWriter() == nil || len(phone.ActionHandlers()) != 1 {
		t.Fatal("a new device needs its config, log writer and action handler")
	}
}

func TestListedDeviceRunsAppActions(t *testing.T) {
	h, runner := newTestHandler(&memoryStore{})
	runner.Script(devicesLine, []byte("emulator-5554 device product:sdk model:Pixel\n"), nil)
	scriptProps(runner, "emulator-5554", "14")
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	devs, _ := h.GetDevices()
	dev := devs[0]
	scriptAll(runner,
		"adb -s emulator-5554 install -t -r -g /apps/game.apk",
		"adb -s emulator-5554 shell am force-stop com.example.game",
		"adb -s emulator-5554 uninstall com.example.game")
	runner.Script("adb -s emulator-5554 shell pm list packages", []byte("package:com.example.game"), nil)

	if err := dev.InstallApp(testApp()); err != nil {
		t.Fatal(err)
	}
	if ok, err := dev.IsAppInstalled(testApp()); err != nil || !ok {
		t.Fatalf("installed = %v, %v", ok, err)
	}
	if err := dev.StopApp(testApp()); err != nil {
		t.Fatal(err)
	}
	if err := dev.UninstallApp(testApp()); err != nil {
		t.Fatal(err)
	}
	got := runner.Commands()
	tail := got[len(got)-4:]
	want := []string{
		"adb -s emulator-5554 install -t -r -g /apps/game.apk",
		"adb -s emulator-5554 shell pm list packages",
		"adb -s emulator-5554 shell am force-stop com.example.game",
		"adb -s emulator-5554 uninstall com.example.game",
	}
	if strings.Join(tail, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands\n got %q\nwant %q", tail, want)
	}
}

func TestApplyDeviceListFollowsAuthorization(t *testing.T) {
	store := &memoryStore{}
	h, runner := newTestHandler(store)
	scriptProps(runner, "RZCTA1YRJ5W", "13")
	state := func() device.State {
		t.Helper()
		d, ok := h.Get("RZCTA1YRJ5W")
		if !ok {
			t.Fatal("device not known")
		}
		return d.DeviceState()
	}

	h.applyDeviceList(parseADBDeviceList("RZCTA1YRJ5W            unauthorized usb:0-1.3 transport_id:2\n"))
	d, _ := h.Get("RZCTA1YRJ5W")
	if state() != device.StateUnauthorized || d.Parameter != nil {
		t.Fatalf("state %v, properties must not be read before authorization: %v", state(), d.Parameter)
	}
	h.applyDeviceList(parseADBDeviceList("RZCTA1YRJ5W            device usb:0-1.3 product:r9q2xeea model:SM_G990B2 device:r9q transport_id:3\n"))
	if state() != device.StateBooted || d.DeviceParameter()["Device Model"] != "SM_G990B2" || d.OSVersion != "13" {
		t.Fatalf("state %v params %v", state(), d.Parameter)
	}
	h.applyDeviceList(parseADBDeviceList("RZCTA1YRJ5W            offline transport_id:3\n"))
	if state() != device.StateRemoteDisconnected {
		t.Fatalf("offline: state %v", state())
	}
	h.applyDeviceList(nil) // unplugged
	if state() != device.StateRemoteDisconnected {
		t.Fatalf("unplugged: state %v", state())
	}
	if len(store.created) != 1 {
		t.Fatalf("the device must be stored once: %d", len(store.created))
	}
}

func TestApplyDeviceListKeepsStoredDevices(t *testing.T) {
	cfg := &models.Device{DeviceIdentifier: "emulator-5554", Manager: Manager}
	store := &memoryStore{devices: models.Devices{cfg}}
	h, runner := newTestHandler(store)
	runner.Script(devicesLine, nil, nil)
	if err := h.Init("http://hub", "node1", nil); err != nil {
		t.Fatal(err)
	}
	d, ok := h.Get("emulator-5554")
	if !ok || d.GetConfig() != cfg || d.DeviceState() != device.StateRemoteDisconnected {
		t.Fatalf("stored device: ok %v state %v", ok, d.DeviceState())
	}
	runner.Script(devicesLine, []byte("emulator-5554\tdevice product:sdk model:Pixel device:emu transport_id:1\n"), nil)
	scriptProps(runner, "emulator-5554", "14")
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	if d.DeviceState() != device.StateBooted || d.GetConfig() != cfg || len(store.created) != 0 {
		t.Fatalf("state %v config %v created %d", d.DeviceState(), d.GetConfig(), len(store.created))
	}
}

func TestInitConnectsAndDisconnectsNetworkDevices(t *testing.T) {
	net := &models.Device{DeviceIdentifier: "10.0.0.5:5555", ConnectionParameter: &models.ConnectionParameter{
		ConnectionType: models.ConnectionTypeRemote, IP: "10.0.0.5", Port: 5555}}
	usb := &models.Device{DeviceIdentifier: "usb1", ConnectionParameter: &models.ConnectionParameter{ConnectionType: models.ConnectionTypeUSB}}
	h, runner := newTestHandler(&memoryStore{devices: models.Devices{net, usb, {DeviceIdentifier: "bare"}}})
	scriptAll(runner, "adb connect 10.0.0.5:5555", "adb disconnect 10.0.0.5:5555")
	runner.Script(devicesLine, []byte("List of devices attached\n"), nil)

	if err := h.Init("http://hub", "node1", nil); err != nil {
		t.Fatal(err)
	}
	assertCommands(t, runner, "adb connect 10.0.0.5:5555", devicesLine, "adb disconnect 10.0.0.5:5555")
	if devs, _ := h.GetDevices(); len(devs) != 3 {
		t.Fatalf("devices = %d", len(devs))
	}
}

func TestInitFailures(t *testing.T) {
	remote := &models.Device{DeviceIdentifier: "10.0.0.5:5555", ConnectionParameter: &models.ConnectionParameter{
		ConnectionType: models.ConnectionTypeRemote, IP: "10.0.0.5", Port: 5555}}
	tests := []struct {
		name   string
		store  *memoryStore
		script func(r *testkit.FakeRunner)
	}{
		{"storage", &memoryStore{getErr: errBoom}, func(*testkit.FakeRunner) {}},
		{"connect", &memoryStore{devices: models.Devices{remote}}, func(*testkit.FakeRunner) {}},
		{"list", &memoryStore{devices: models.Devices{remote}}, func(r *testkit.FakeRunner) { scriptAll(r, "adb connect 10.0.0.5:5555") }},
		{"disconnect", &memoryStore{devices: models.Devices{remote}}, func(r *testkit.FakeRunner) {
			scriptAll(r, "adb connect 10.0.0.5:5555", devicesLine)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, runner := newTestHandler(tt.store)
			tt.script(runner)
			if err := h.Init("http://hub", "node1", nil); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestStorageErrorsDoNotStopTheRefresh(t *testing.T) {
	h, runner := newTestHandler(&memoryStore{createErr: errBoom, updateErr: errBoom})
	runner.Script(devicesLine, []byte("R58M1234 unauthorized\n"), nil)
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	if !h.Has("R58M1234") {
		t.Fatal("device missing")
	}
}

func TestRefreshDevicesFailure(t *testing.T) {
	h, _ := newTestHandler(&memoryStore{})
	if err := h.RefreshDevices(true); err == nil {
		t.Fatal("expected an error when adb fails")
	}
}

func TestRefreshDevicesIsASafetyNetWhileTracking(t *testing.T) {
	h, runner := newTestHandler(&memoryStore{})
	clock := testkit.NewFakeClock(time.Unix(1000, 0))
	h.deps.Clock = clock
	runner.Script(devicesLine, nil, nil)
	h.tracking.Store(true)

	for _, step := range []struct {
		advance time.Duration
		force   bool
		polls   int
	}{
		{0, false, 1},
		{30 * time.Second, false, 1},
		{31 * time.Second, false, 2},
		{0, true, 3},
	} {
		clock.Advance(step.advance)
		if err := h.RefreshDevices(step.force); err != nil {
			t.Fatal(err)
		}
		if got := len(runner.Commands()); got != step.polls {
			t.Fatalf("after %+v: %d polls, want %d", step, got, step.polls)
		}
	}
	h.tracking.Store(false)
	_ = h.RefreshDevices(false)
	if got := len(runner.Commands()); got != 4 {
		t.Fatalf("without tracking every refresh polls: %d", got)
	}
}

func TestStartWatchesAdbAndStopKillsTheServer(t *testing.T) {
	srv := newFakeADBServer(t, func(conn net.Conn, _ string) {
		_, _ = conn.Write([]byte("OKAY"))
		writeADBMessage(conn, "R58M1234\tunauthorized\n")
		time.Sleep(time.Second)
	})
	h, runner := newTestHandler(&memoryStore{})
	h.adbServerAddr = srv.addr()
	scriptAll(runner, "adb start-server", "adb kill-server")
	var changes atomic.Int32
	h.SetChangeNotifier(func() { changes.Add(1) })

	if err := h.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for changes.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if changes.Load() == 0 || !h.Has("R58M1234") || !h.tracking.Load() {
		t.Fatalf("changes %d tracking %v", changes.Load(), h.tracking.Load())
	}
	if err := h.Stop(); err != nil {
		t.Fatal(err)
	}
	if h.stopWatch != nil {
		t.Fatal("the watch must be stopped")
	}
	if err := h.Stop(); err != nil { // stopping twice is fine
		t.Fatal(err)
	}
}

func TestStartReportsAFailingAdbServer(t *testing.T) {
	h, _ := newTestHandler(&memoryStore{})
	h.adbServerAddr = "127.0.0.1:1" // nothing listens
	err := h.Start()
	h.stopWatch() // ends the reconnect loop of the watcher
	if err == nil {
		t.Fatal("expected the start-server error")
	}
}

func TestStartAndStopDevice(t *testing.T) {
	remote := &models.Device{DeviceIdentifier: "10.0.0.5:5555", ConnectionParameter: &models.ConnectionParameter{
		ConnectionType: models.ConnectionTypeRemote, IP: "10.0.0.5", Port: 5555}}
	usb := &models.Device{DeviceIdentifier: "usb1", ConnectionParameter: &models.ConnectionParameter{ConnectionType: models.ConnectionTypeUSB}}
	store := &memoryStore{devices: models.Devices{remote, usb}}
	h, runner := newTestHandler(store)
	scriptAll(runner, "adb connect 10.0.0.5:5555", "adb disconnect 10.0.0.5:5555", devicesLine)
	if err := h.Init("http://hub", "node1", nil); err != nil {
		t.Fatal(err)
	}
	before := len(runner.Commands())

	for _, id := range []string{"10.0.0.5:5555", "usb1"} {
		if err := h.StartDevice(id); err != nil {
			t.Fatalf("StartDevice(%s): %v", id, err)
		}
	}
	booted, _ := h.Get("usb1")
	booted.State = device.StateBooted
	if err := h.StartDevice("usb1"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"10.0.0.5:5555", "usb1"} {
		if err := h.StopDevice(id); err != nil {
			t.Fatalf("StopDevice(%s): %v", id, err)
		}
	}
	assertCommands(t, runner, append(runner.Commands()[:before], "adb connect 10.0.0.5:5555", "adb disconnect 10.0.0.5:5555")...)
}

func TestStartAndStopDeviceFailures(t *testing.T) {
	h, _ := newTestHandler(&memoryStore{})
	if err := h.StartDevice("nope"); err != device.DeviceNotFoundError {
		t.Fatalf("StartDevice error = %v", err)
	}
	if err := h.StopDevice("nope"); err != device.DeviceNotFoundError {
		t.Fatalf("StopDevice error = %v", err)
	}
	// a device that is not in the storage can't be started
	d := newDevice(h.deps, h.video, "ghost")
	d.State = device.StateRemoteDisconnected
	h.Add(d)
	if err := h.StartDevice("ghost"); err != device.DeviceNotFoundError {
		t.Fatalf("ghost error = %v", err)
	}
}

func TestRegisterDeviceIsUnsupported(t *testing.T) {
	h, _ := newTestHandler(&memoryStore{})
	if _, err := h.RegisterDevice(device.RegisterData{}); err == nil || !strings.Contains(err.Error(), device.ErrUnsupported.Error()) {
		t.Fatalf("error = %v", err)
	}
	if h.Name() != Manager {
		t.Fatal("manager name")
	}
}

func TestNewHandlerUsesTheRealRunner(t *testing.T) {
	h := NewHandler(&memoryStore{})
	if _, ok := h.deps.Runner.(device.ExecRunner); !ok {
		t.Fatalf("runner = %T", h.deps.Runner)
	}
	if _, ok := h.deps.Clock.(device.RealClock); !ok || h.deps.NewID() == "" || h.deps.AAPT == nil {
		t.Fatal("defaults are missing")
	}
}
