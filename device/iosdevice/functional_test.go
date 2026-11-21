package iosdevice

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// TestHandlerFunctionalRefreshStartAndAppActions walks through a hub's use of
// an iPhone with fake go-ios output, a fake xcodebuild and a fake
// WebDriverAgent: refresh finds the device, StartDevice starts WebDriverAgent
// and the app actions reach it.
func TestHandlerFunctionalRefreshStartAndAppActions(t *testing.T) {
	r := newRig(t, "00008110-0001")

	// the device list shows the phone
	if err := r.h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	devices, err := r.h.GetDevices()
	if err != nil || len(devices) != 1 {
		t.Fatalf("GetDevices = %v, %v", devices, err)
	}
	phone := devices[0]
	if phone.DeviceID() != "00008110-0001" || phone.DeviceState() != device.StateBooted || phone.DeviceName() != "iPhone of 00008110-0001" {
		t.Fatalf("phone = %s %v %q", phone.DeviceID(), phone.DeviceState(), phone.DeviceName())
	}

	// WebDriverAgent starts through xcodebuild, with the exact command line
	if err := r.h.StartDevice(phone.DeviceID()); err != nil {
		t.Fatal(err)
	}
	wantCommand := "xcodebuild test-without-building -xctestrun " + xctestrun + " -destination id=00008110-0001"
	if got := r.proc.runner.Commands(); !reflect.DeepEqual(got, []string{wantCommand}) {
		t.Fatalf("commands = %q", got)
	}
	r.wda.saw(t, `POST /session {"capabilities":{}}`)

	// the app actions go through WebDriverAgent
	params := &app.Parameter{Identifier: "com.example.game"}
	if err := phone.StartApp(nil, params, "session-1", "http://node"); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/apps/launch {"arguments":["SESSION_ID","session-1","DEVICE_ID","00008110-0001","NODE_URL","http://node"]`)
	input, ok := device.As[device.InputCapable](phone)
	if !ok {
		t.Fatal("an iOS device must be InputCapable")
	}
	if err := input.Tap(100, 200); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, `POST /session/s1/wda/tap {"x":50,"y":100}`)
	if err := phone.StopApp(&app.Parameter{Identifier: "com.app"}); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, "POST /session/s1/wda/apps/terminate")

	// stopping ends the session and xcodebuild
	if err := r.h.StopDevice(phone.DeviceID()); err != nil {
		t.Fatal(err)
	}
	r.wda.saw(t, "DELETE /session")
	waitFor(t, func() bool { return len(r.proc.proc.got()) == 1 })
	if got := r.proc.proc.got(); got[0] != "interrupt" {
		t.Errorf("signals = %v", got)
	}

	// unplugging the phone shows in the next refresh
	r.clock.Advance(1)
	r.usb.serials = nil
	_ = r.h.RefreshDevices(true)
	if phone.DeviceState() != device.StateRemoteDisconnected {
		t.Errorf("state after unplug = %v", phone.DeviceState())
	}
}
