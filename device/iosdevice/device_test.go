package iosdevice

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// startedDevice is a device with a running WebDriverAgent session.
func startedDevice(t *testing.T) (*Device, *rig) {
	t.Helper()
	r := newRig(t, "U")
	if err := r.h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	if err := r.h.StartDevice("U"); err != nil {
		t.Fatal(err)
	}
	return r.device(t, "U"), r
}

func TestNewDeviceIdentity(t *testing.T) {
	d := newDevice("U1", nil, Deps{}.withDefaults())
	if d.DeviceID() != "U1" || d.DeviceType() != int(models.DeviceTypePhone) || d.PlatformType() != int(models.PlatformTypeiOS) {
		t.Errorf("identity = %s %d %d", d.DeviceID(), d.DeviceType(), d.PlatformType())
	}
	var response ios.GetAllValuesResponse
	response.Value.DeviceName, response.Value.ProductName = "Phone", "iPhone OS"
	response.Value.ProductVersion, response.Value.ProductType = "18.1", "iPhone15,2"
	d.UpdateDeviceInfos(response)
	if d.DeviceName() != "Phone" || d.DeviceOSName() != "iPhone OS" || d.DeviceOSVersion() != "18.1" || d.DeviceParameter()["Device Model"] != "iPhone15,2" {
		t.Errorf("infos = %s %s %s %v", d.DeviceName(), d.DeviceOSName(), d.DeviceOSVersion(), d.DeviceParameter())
	}
	if d.TargetVersion() != "" || d.ConnectionTimeout() != ConnectionTimeout || d.IsAppConnected() {
		t.Error("defaults differ")
	}
}

func TestRoles(t *testing.T) {
	d := newDevice("U", nil, Deps{}.withDefaults())
	want := []string{"Screenshotter", "NativeScripter", "InputCapable", "VideoStreamer"}
	if got := testkit.Roles(d); !reflect.DeepEqual(got, want) {
		t.Errorf("roles = %v, want %v", got, want)
	}
}

func TestAppCallsNeedAWebDriverSession(t *testing.T) {
	d := newDevice("U", nil, Deps{}.withDefaults())
	log := &recordingWriter{}
	d.SetLogWriter(log)
	params := &app.Parameter{Identifier: "com.app"}
	if err := d.StartApp(nil, params, "s", "n"); err == nil {
		t.Error("StartApp without session should fail")
	}
	if err := d.StopApp(params); err == nil {
		t.Error("StopApp without session should fail")
	}
	d.RunNativeScript([]byte("log('x')"))
	if len(log.errors) != 1 || !strings.Contains(log.errors[0], "no WebDriverAgent session") {
		t.Errorf("script errors = %q", log.errors)
	}
}

func TestStartAndStopApp(t *testing.T) {
	d, r := startedDevice(t)
	if err := d.StartApp(nil, &app.Parameter{Identifier: "com.app"}, "sess", "http://node"); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/apps/launch {"arguments":["SESSION_ID","sess","DEVICE_ID","U","NODE_URL","http://node"],"bundleId":"com.app","shouldWaitForQuiescence":1}`)
	if err := d.StopApp(&app.Parameter{Identifier: "com.app"}); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, "POST /session/s1/wda/apps/terminate")
	before := len(r.wda.requests())
	if err := d.StopApp(&app.Parameter{Identifier: "com.other"}); err != nil {
		t.Fatal(err)
	}
	if got := r.wda.requests()[before:]; len(got) != 1 || !strings.HasPrefix(got[0], "GET /session/s1/wda/activeAppInfo") {
		t.Errorf("a different foreground app must not be terminated: %q", got)
	}
}

func TestStopAppReportsWebDriverErrors(t *testing.T) {
	d, r := startedDevice(t)
	r.wda.Close()
	if err := d.StopApp(&app.Parameter{Identifier: "com.app"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRunNativeScriptFailureGoesHomeAndClosesTheSession(t *testing.T) {
	d, r := startedDevice(t)
	log := &recordingWriter{}
	d.SetLogWriter(log)
	d.RunNativeScript([]byte("this is not lua ("))
	if len(log.errors) == 0 || !strings.Contains(log.errors[0], "Script Failed") {
		t.Fatalf("errors = %q", log.errors)
	}
	r.wda.saw(t, `POST /session/s1/wda/pressButton {"name":"home"}`)
	r.wda.saw(t, "DELETE /session")
}

func TestRunNativeScriptRunsAScript(t *testing.T) {
	d, r := startedDevice(t)
	d.RunNativeScript([]byte(`hub.click("name", "Login")`))
	r.wda.saw(t, "POST /session/s1/element/e1/click")
}

func TestInputActionsScaleScreenshotPixelsToPoints(t *testing.T) {
	d, r := startedDevice(t)
	if err := d.Tap(100, 200); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/tap {"x":50,"y":100}`)
	if err := d.Swipe(0, 0, 100, 200, 0); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/dragfromtoforduration {"duration":0.3,"fromX":0,"fromY":0,"toX":50,"toY":100}`)
	if err := d.InputText("hi"); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/keys {"value":["h","i"]}`)
	for key, button := range map[string]string{"VOLUME_UP": "volumeUp", "volume_down": "volumeDown", "home": "home"} {
		if err := d.PressKey(key); err != nil {
			t.Fatal(err)
		}
		r.wda.saw(t, `POST /session/s1/wda/pressButton {"name":"`+button+`"}`)
	}
	if err := d.PressKey("back"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("PressKey(back) = %v", err)
	}
}

func TestInputActionsNeedASession(t *testing.T) {
	d := newDevice("U", nil, Deps{}.withDefaults())
	for name, err := range map[string]error{
		"Tap":       d.Tap(1, 1),
		"Swipe":     d.Swipe(1, 1, 2, 2, time.Second),
		"InputText": d.InputText("x"),
		"PressKey":  d.PressKey("home"),
	} {
		if err == nil {
			t.Errorf("%s without session should fail", name)
		}
	}
}

func TestStartXCUITestRunnerReusesARunningWebDriverAgent(t *testing.T) {
	d, r := startedDevice(t)
	if err := d.StartXCUITestRunner(); err != nil {
		t.Fatal(err)
	}
	if n := len(r.proc.runner.Commands()); n != 1 {
		t.Errorf("xcodebuild started %d times", n)
	}
}

func TestStartXCUITestRunnerRestartsARunnerThatStoppedAnswering(t *testing.T) {
	d, r := startedDevice(t)
	r.wda.Close()
	r.proc.output = wdaURLLine
	other := newWDAServer(t)
	r.usb.forwardTo = other.hostPort()
	if err := d.StartXCUITestRunner(); err != nil {
		t.Fatal(err)
	}
	if n := len(r.proc.runner.Commands()); n != 2 {
		t.Errorf("xcodebuild started %d times", n)
	}
	other.saw(t, "POST /session")
}

func TestStartXCUITestRunnerReportsSessionFailure(t *testing.T) {
	d, r := startedDevice(t)
	r.wda.remove("POST /session")
	err := d.StartXCUITestRunner() // runs, so it only opens a new session
	if err == nil || !strings.Contains(err.Error(), "create WebDriverAgent session") {
		t.Fatalf("err = %v", err)
	}
}

func TestStartXCUITestRunnerReportsAttachAndStartFailure(t *testing.T) {
	r := newRigWith(t, nil, "U")
	_ = r.h.RefreshDevices(true)
	err := r.h.StartDevice("U")
	if err == nil || !strings.Contains(err.Error(), "no running WebDriverAgent (nothing listens), and starting one failed") {
		t.Fatalf("err = %v", err)
	}
}

func TestStartXCUITestRunnerUsesARunningWebDriverAgentOnTheDevice(t *testing.T) {
	r := newRig(t, "U")
	r.usb.probeErr = nil
	_ = r.h.RefreshDevices(true)
	if err := r.h.StartDevice("U"); err != nil {
		t.Fatal(err)
	}
	if n := len(r.proc.runner.Commands()); n != 0 {
		t.Errorf("xcodebuild started %d times", n)
	}
	if err := r.h.StopDevice("U"); err != nil {
		t.Fatal(err)
	}
	if r.usb.closed != 2 {
		t.Errorf("closed %d forwards", r.usb.closed)
	}
}

func TestStopXCUITestRunnerClosesTheSessionAndStopsXcodebuild(t *testing.T) {
	d, r := startedDevice(t)
	if err := d.StopXCUITestRunner(); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, "DELETE /session")
	waitFor(t, func() bool { return len(r.proc.proc.got()) == 1 })
	if d.webDriver != nil || d.wdaRunner != nil {
		t.Error("the runner is still set")
	}
	if err := d.StopXCUITestRunner(); err != nil { // stopping twice is fine
		t.Fatal(err)
	}
}

func TestEnsureWebDriver(t *testing.T) {
	d, r := startedDevice(t)
	wd, mjpeg, err := d.ensureWebDriver()
	if err != nil || wd != d.webDriver || mjpeg != "http://127.0.0.1:9100/" {
		t.Fatalf("ensureWebDriver = %v, %q, %v", wd, mjpeg, err)
	}
	// without a runner the MJPEG server is on the host of WebDriverAgent
	d.wdaRunner = nil
	if _, mjpeg, err = d.ensureWebDriver(); err != nil || !strings.HasSuffix(mjpeg, ":9100/") {
		t.Fatalf("ensureWebDriver = %q, %v", mjpeg, err)
	}
	// a stopped server is started again
	r.wda.Close()
	d.wdaRunner = nil
	r.usb.serials = nil
	if _, _, err := d.ensureWebDriver(); err == nil {
		t.Fatal("expected a start failure")
	}
}

func TestStartVideoNeedsABootedDevice(t *testing.T) {
	d := newDevice("U", nil, Deps{}.withDefaults())
	if _, err := d.StartVideo(context.Background(), device.VideoOptions{}); err == nil {
		t.Fatal("expected an error")
	}
	d.State = device.StateBooted
	d.wda.cfg = &config.WebDriver{}
	d.wda.deps.USB.GetDevice = func(string) (ios.DeviceEntry, error) { return ios.DeviceEntry{}, errors.New("gone") }
	d.wda.deps.USB.Probe = func(ios.DeviceEntry, uint16) error { return errors.New("none") }
	if _, err := d.StartVideo(context.Background(), device.VideoOptions{}); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}
