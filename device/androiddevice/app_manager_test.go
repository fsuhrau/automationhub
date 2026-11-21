package androiddevice

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func testApp() *app.Parameter {
	return &app.Parameter{
		Identifier: "com.example.game",
		App: &app.AppParams{
			AppPath: "/apps/game.apk",
			Hash:    "h1",
			Android: &app.AndroidParams{LaunchActivity: ".Main"},
		},
	}
}

func TestInstallAppCommands(t *testing.T) {
	tests := []struct {
		name  string
		api   int64
		aapt  func() (string, error)
		badge string
		want  string
	}{
		{name: "modern android", api: 30, want: "adb -s S1 install -t -r -g /apps/game.apk"},
		{name: "old android, release build", api: 21, badge: "package: name='x'", want: "adb -s S1 install -t -rg /apps/game.apk"},
		{name: "old android, debuggable build", api: 21, badge: "application-debuggable", want: "adb -s S1 install -t -rgd /apps/game.apk"},
		{name: "old android, no aapt", api: 21, aapt: func() (string, error) { return "", errBoom }, want: "adb -s S1 install -t -rg /apps/game.apk"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &testkit.FakeRunner{}
			runner.Script(tt.want, nil, nil)
			runner.Script("aapt dump badging /apps/game.apk", []byte(tt.badge), nil)
			d := newTestDevice(runner)
			d.apiLevel = tt.api
			if tt.aapt != nil {
				d.deps.AAPT = tt.aapt
			}
			if err := d.InstallApp(testApp()); err != nil {
				t.Fatal(err)
			}
			if d.installedApps["com.example.game"] != "h1" {
				t.Fatalf("installed apps = %v", d.installedApps)
			}
			if got := runner.Commands(); got[len(got)-1] != tt.want {
				t.Fatalf("commands = %q, want last %q", got, tt.want)
			}
		})
	}
}

func TestInstallAppFailures(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 install -t -r -g /apps/game.apk", nil, errBoom)
	d := newTestDevice(runner)
	d.apiLevel = 30
	if err := d.InstallApp(testApp()); err == nil || !strings.Contains(err.Error(), "installation failed") {
		t.Fatalf("error = %v", err)
	}
	if len(d.installedApps) != 0 {
		t.Fatalf("installed apps = %v", d.installedApps)
	}
	// aapt that fails to run counts as not debuggable
	runner.Script("aapt dump badging /apps/game.apk", nil, errBoom)
	if d.isDebuggable("/apps/game.apk") {
		t.Fatal("a failing aapt must not report debuggable")
	}
}

func TestUninstallApp(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 uninstall com.example.game", nil, nil)
	d := newTestDevice(runner)
	d.installedApps["com.example.game"] = "h1"
	if err := d.UninstallApp(testApp()); err != nil || d.installedApps["com.example.game"] != "" {
		t.Fatalf("err %v apps %v", err, d.installedApps)
	}
	failing := newTestDevice(&testkit.FakeRunner{})
	failing.installedApps["com.example.game"] = "h1"
	if err := failing.UninstallApp(testApp()); err == nil || failing.installedApps["com.example.game"] != "h1" {
		t.Fatalf("a failed uninstall must keep the record: %v %v", err, failing.installedApps)
	}
}

func TestIsAppInstalled(t *testing.T) {
	tests := []struct {
		name     string
		packages string
		known    string
		err      error
		want     bool
	}{
		{name: "installed and same build", packages: "package:com.example.game\n", known: "h1", want: true},
		{name: "installed, other build", packages: "package:com.example.game\n", known: "h0"},
		{name: "installed, never installed here", packages: "package:com.example.game\n"},
		{name: "not installed", packages: "package:other\n", known: "h1"},
		{name: "adb fails", err: errBoom, known: "h1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &testkit.FakeRunner{}
			runner.Script("adb -s S1 shell pm list packages", []byte(tt.packages), tt.err)
			d := newTestDevice(runner)
			if tt.known != "" {
				d.installedApps["com.example.game"] = tt.known
			}
			got, err := d.IsAppInstalled(testApp())
			if got != tt.want || err != tt.err {
				t.Fatalf("installed = %v, err = %v", got, err)
			}
		})
	}
}

func TestStopAndClearApp(t *testing.T) {
	runner := &testkit.FakeRunner{}
	scriptAll(runner, "adb -s S1 shell am force-stop com.example.game", "adb -s S1 shell pm clear com.example.game")
	d := newTestDevice(runner)
	if err := d.StopApp(testApp()); err != nil {
		t.Fatal(err)
	}
	if err := d.ClearAppData(testApp()); err != nil {
		t.Fatal(err)
	}
	assertCommands(t, runner, "adb -s S1 shell am force-stop com.example.game", "adb -s S1 shell pm clear com.example.game")
}

// unlockedScript scripts an awake, unlocked device.
func unlockedScript(runner *testkit.FakeRunner) {
	runner.Script("adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker", []byte("mHoldingDisplaySuspendBlocker=true"), nil)
	runner.Script("adb -s S1 shell dumpsys window | grep mDreamingLockscreen", []byte("mDreamingLockscreen=false"), nil)
}

func startLine(url string) string {
	return "adb -s S1 shell am start -n com.example.game/.Main -e SESSION_ID sess -e NODE_URL " + url + " -e DEVICE_ID S1"
}

func TestStartAppReversesLoopbackNodes(t *testing.T) {
	runner := &testkit.FakeRunner{}
	unlockedScript(runner)
	scriptAll(runner, "adb -s S1 reverse tcp:8004 tcp:8004", startLine("ws://localhost:8004/device/connect"))
	d := newTestDevice(runner)
	err := d.StartApp(&device.DeviceConfig{}, testApp(), "sess", "ws://localhost:8004/device/connect")
	if err != nil {
		t.Fatal(err)
	}
	assertCommands(t, runner,
		"adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker",
		"adb -s S1 shell dumpsys window | grep mDreamingLockscreen",
		"adb -s S1 reverse tcp:8004 tcp:8004",
		startLine("ws://localhost:8004/device/connect"))
}

func TestStartAppKeepsRemoteNodes(t *testing.T) {
	runner := &testkit.FakeRunner{}
	unlockedScript(runner)
	runner.Script(startLine("ws://hub:8004/c"), nil, nil)
	if err := newTestDevice(runner).StartApp(&device.DeviceConfig{}, testApp(), "sess", "ws://hub:8004/c"); err != nil {
		t.Fatal(err)
	}
	if got := runner.Commands(); strings.Contains(strings.Join(got, "\n"), "reverse") {
		t.Fatalf("a remote node needs no reverse: %q", got)
	}
}

func TestStartAppFailures(t *testing.T) {
	cfg := &device.DeviceConfig{}
	t.Run("unlock fails", func(t *testing.T) {
		if err := newTestDevice(&testkit.FakeRunner{}).StartApp(cfg, testApp(), "sess", "ws://hub/c"); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("reverse fails", func(t *testing.T) {
		runner := &testkit.FakeRunner{}
		unlockedScript(runner)
		err := newTestDevice(runner).StartApp(cfg, testApp(), "sess", "ws://localhost:8004/c")
		if err == nil || !strings.Contains(err.Error(), "adb reverse tcp:8004") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("am start fails", func(t *testing.T) {
		runner := &testkit.FakeRunner{}
		unlockedScript(runner)
		if err := newTestDevice(runner).StartApp(cfg, testApp(), "sess", "ws://hub/c"); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestStartAppWakesAndUnlocksTheDevice(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker", []byte("mHoldingDisplaySuspendBlocker=false"), nil)
	runner.Script("adb -s S1 shell dumpsys window | grep mDreamingLockscreen", []byte("mDreamingLockscreen=true"), nil)
	key := func(code int) string { return fmt.Sprintf("adb -s S1 shell input keyevent %d", code) }
	lines := []string{
		key(KEYCODE_WAKEUP), key(KEYCODE_MENU), "adb -s S1 shell input swipe 400 800 400 200",
		key(KEYCODE_NUMPAD_0 + 1), key(KEYCODE_NUMPAD_0 + 2), key(KEYCODE_NUMPAD_ENTER),
		startLine("ws://hub/c"),
	}
	scriptAll(runner, lines...)
	cfg := &device.DeviceConfig{DeviceParameter: map[string]string{"PIN": "12"}}
	if err := newTestDevice(runner).StartApp(cfg, testApp(), "sess", "ws://hub/c"); err != nil {
		t.Fatal(err)
	}
	want := append([]string{
		"adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker",
	}, lines[:3]...)
	want = append(want, "adb -s S1 shell dumpsys window | grep mDreamingLockscreen")
	want = append(want, lines[3:]...)
	assertCommands(t, runner, want...)
}
