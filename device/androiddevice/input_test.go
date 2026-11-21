package androiddevice

import (
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestInputCommands(t *testing.T) {
	tests := []struct {
		name string
		call func(d *Device) error
		want string
	}{
		{"tap", func(d *Device) error { return d.Tap(10, 20) }, "adb -s S1 shell input tap 10 20"},
		{"swipe default duration", func(d *Device) error { return d.Swipe(1, 2, 3, 4, 0) }, "adb -s S1 shell input swipe 1 2 3 4"},
		{"swipe with duration", func(d *Device) error { return d.Swipe(1, 2, 3, 4, 250*time.Millisecond) }, "adb -s S1 shell input swipe 1 2 3 4 250"},
		{"text", func(d *Device) error { return d.InputText("a b") }, "adb -s S1 shell input text a%sb"},
		{"named key", func(d *Device) error { return d.PressKey("HOME") }, "adb -s S1 shell input keyevent 3"},
		{"numeric key", func(d *Device) error { return d.PressKey("66") }, "adb -s S1 shell input keyevent 66"},
		{"send text", func(d *Device) error { return d.sendText("x") }, "adb -s S1 shell input text x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &testkit.FakeRunner{}
			runner.Script(tt.want, nil, nil)
			if err := tt.call(newTestDevice(runner)); err != nil {
				t.Fatal(err)
			}
			assertCommands(t, runner, tt.want)
		})
	}
}

func TestInputWithoutEffect(t *testing.T) {
	runner := &testkit.FakeRunner{}
	d := newTestDevice(runner)
	if err := d.InputText(""); err != nil {
		t.Fatal(err)
	}
	if err := d.PressKey("nope"); err == nil {
		t.Fatal("expected an error for an unknown key")
	}
	assertCommands(t, runner)
}

func TestInputFailuresAreReturned(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{})
	calls := map[string]error{
		"tap":     d.Tap(1, 1),
		"swipe":   d.Swipe(1, 1, 1, 1, 0),
		"text":    d.InputText("x"),
		"key":     d.PressKey("home"),
		"keycode": d.pressKey(1),
	}
	for name, err := range calls {
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestNamedKeyIsCaseInsensitive(t *testing.T) {
	if code, ok := namedKey("Volume_Up"); !ok || code != KEYCODE_VOLUME_UP {
		t.Fatalf("code %d ok %v", code, ok)
	}
	if _, ok := namedKey("x"); ok {
		t.Fatal("unknown key found")
	}
}

func TestReadLogs(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 logcat -d -v time -t 200", []byte("a\r\nb\r\n"), nil)
	runner.Script("adb -s S1 logcat -d -v time -t 5", nil, nil)
	d := newTestDevice(runner)
	lines, err := d.ReadLogs(0)
	if err != nil || strings.Join(lines, "|") != "a|b" {
		t.Fatalf("lines %q err %v", lines, err)
	}
	if lines, err := d.ReadLogs(5); err != nil || lines != nil {
		t.Fatalf("empty log: %q %v", lines, err)
	}
	if _, err := d.ReadLogs(7); err == nil {
		t.Fatal("expected an error")
	}
}

func TestReadLogsSince(t *testing.T) {
	const line = "adb -s S1 logcat -d -v time -T 01-02 03:04:05.678"
	runner := &testkit.FakeRunner{}
	runner.Script(line, []byte("1\n2\n3\n"), nil)
	d := newTestDevice(runner)
	lines, err := d.ReadLogsSince("01-02 03:04:05.678", 2)
	if err != nil || strings.Join(lines, "|") != "2|3" {
		t.Fatalf("lines %q err %v", lines, err)
	}
	if lines, err := d.ReadLogsSince("01-02 03:04:05.678", 0); err != nil || len(lines) != 3 {
		t.Fatalf("default limit: %q %v", lines, err)
	}
	if _, err := d.ReadLogsSince("yesterday", 1); err == nil {
		t.Fatal("expected an error for an invalid time")
	}
	if _, err := d.ReadLogsSince("01-02 03:04:05.679", 1); err == nil {
		t.Fatal("expected an error from adb")
	}
}

func TestFeatures(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell input keyevent 4", nil, nil)
	d := newTestDevice(runner)
	if !d.HasFeature("back") || d.HasFeature("home") {
		t.Fatal("only back is a feature")
	}
	d.Execute("back")
	d.Execute("other")
	assertCommands(t, runner, "adb -s S1 shell input keyevent 4")
}

func TestRunNativeScript(t *testing.T) {
	runner := &testkit.FakeRunner{}
	d := newTestDevice(runner)
	d.RunNativeScript([]byte(`local x = 1`))
	assertCommands(t, runner)

	runner.Script("adb -s S1 shell input keyevent 4", nil, nil)
	d.RunNativeScript([]byte(`error("script broke")`))
	assertCommands(t, runner, "adb -s S1 shell input keyevent 4")
}

func TestDeviceIdentity(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{})
	d.SetDeviceState("StateUnauthorized")
	if d.DeviceID() != "S1" || d.DeviceOSName() != "android" || d.DeviceState() != device.StateUnauthorized ||
		d.DeviceType() != int(models.DeviceTypePhone) || d.PlatformType() != int(models.PlatformTypeAndroid) || d.ConnectionTimeout() != ConnectionTimeout || d.IsAppConnected() {
		t.Fatalf("identity %+v", d.Info)
	}
}
