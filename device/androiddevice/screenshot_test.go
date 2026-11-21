package androiddevice

import (
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	execOutLine = "adb -s S1 exec-out screencap -p"
	shotLine    = "adb -s S1 shell screencap -p /sdcard/tmp1.png"
	shotPull    = "adb -s S1 pull /sdcard/tmp1.png tmp1.png"
	shotRm      = "adb -s S1 shell rm /sdcard/tmp1.png"
)

func TestNotReady(t *testing.T) {
	tests := []struct {
		state device.State
		want  string
	}{
		{device.StateBooted, ""},
		{device.StateRemoteDisconnected, "S1 is disconnected"},
		{device.StateUnauthorized, "S1 is unauthorized"},
		{device.StateShutdown, "S1 is shutdown"},
	}
	for _, tt := range tests {
		d := newTestDevice(&testkit.FakeRunner{})
		d.State = tt.state
		err := d.notReady()
		if (err == nil) != (tt.want == "") || (err != nil && !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("state %v: error = %v, want %q", tt.state, err, tt.want)
		}
	}
}

func TestGetScreenshotThroughExecOut(t *testing.T) {
	runner := &testkit.FakeRunner{}
	img := pngBytes(t, 4, 6)
	runner.Script(execOutLine, img, nil)
	data, w, h, err := newTestDevice(runner).GetScreenshot()
	if err != nil || w != 4 || h != 6 || len(data) != len(img) {
		t.Fatalf("w=%d h=%d err=%v", w, h, err)
	}
	assertCommands(t, runner, execOutLine)
}

func TestGetScreenshotRefusesAnUnusableDevice(t *testing.T) {
	runner := &testkit.FakeRunner{}
	d := newTestDevice(runner)
	d.State = device.StateUnauthorized
	if _, _, _, err := d.GetScreenshot(); err == nil {
		t.Fatal("expected an error")
	}
	assertCommands(t, runner)
}

func TestGetScreenshotFallsBackToAFileTransfer(t *testing.T) {
	for name, execOut := range map[string][]byte{"exec-out fails": nil, "exec-out is no png": []byte("garbage")} {
		t.Run(name, func(t *testing.T) {
			inTempDir(t)
			writeFile(t, "tmp1.png", pngBytes(t, 3, 5))
			runner := &testkit.FakeRunner{}
			var execErr error
			if execOut == nil {
				execErr = errBoom
			}
			runner.Script(execOutLine, execOut, execErr)
			scriptAll(runner, shotLine, shotPull, shotRm)
			_, w, h, err := newTestDevice(runner).GetScreenshot()
			if err != nil || w != 3 || h != 5 {
				t.Fatalf("w=%d h=%d err=%v", w, h, err)
			}
			assertCommands(t, runner, execOutLine, shotLine, shotPull, shotRm)
		})
	}
}

func TestGetScreenshotFailures(t *testing.T) {
	tests := []struct {
		name   string
		script func(r *testkit.FakeRunner)
		want   string
	}{
		{"screencap fails", func(r *testkit.FakeRunner) {}, "unscripted"},
		{"pull fails", func(r *testkit.FakeRunner) { scriptAll(r, shotLine, shotRm) }, "unscripted"},
		{"file is no png", func(r *testkit.FakeRunner) { scriptAll(r, shotLine, shotPull, shotRm) }, "decode screenshot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inTempDir(t)
			writeFile(t, "tmp1.png", []byte("garbage"))
			runner := &testkit.FakeRunner{}
			tt.script(runner)
			if _, _, _, err := newTestDevice(runner).GetScreenshot(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestScreencapViaFileMissingFile(t *testing.T) {
	inTempDir(t)
	runner := &testkit.FakeRunner{}
	scriptAll(runner, shotLine, shotPull, shotRm)
	if _, err := newTestDevice(runner).screencapViaFile(); err == nil {
		t.Fatal("expected a read error for the missing file")
	}
}

func TestGetScreenXml(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 exec-out uiautomator dump /dev/tty", []byte(`<hierarchy><node name="a"/></hierarchy>`), nil)
	d := newTestDevice(runner)
	node, err := d.getScreenXml()
	if err != nil || node == nil {
		t.Fatalf("node %v err %v", node, err)
	}
	if _, err := newTestDevice(&testkit.FakeRunner{}).getScreenXml(); err == nil {
		t.Fatal("expected an error")
	}
}
