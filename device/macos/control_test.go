package macos

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/device"
)

const exe = "Contents/MacOS/Game"

func startLine(e *testEnv) string {
	return filepath.Join(e.deps.InstallDir, "h1", "Game.app", exe) + " --sessionId=s1 NODE_URL --nodeURL=n:1 --deviceId=" + testUUID
}

func TestStartAppRunsTheExecutable(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps}
	d.ID = testUUID
	e.runner.Script(startLine(e), nil, nil)

	if err := d.StartApp(nil, writeBundle(t), "s1", "n:1"); err != nil {
		t.Fatal(err)
	}
	if got := e.runner.Commands(); len(got) != 1 || got[0] != startLine(e) {
		t.Fatalf("commands %v", got)
	}
	if d.runningExecutable != "Game" {
		t.Fatalf("runningExecutable %q", d.runningExecutable)
	}
}

func TestStartAppFailure(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps}
	if err := d.StartApp(nil, writeBundle(t), "s1", "n:1"); err == nil {
		t.Fatal("an unscripted start must fail")
	}
}

func TestStopAppKillsByName(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps, runningExecutable: "Game"}
	e.runner.Script("killall Game", nil, nil)

	if err := d.StopApp(writeBundle(t)); err != nil {
		t.Fatal(err)
	}
	if got := e.runner.Commands(); len(got) != 1 || got[0] != "killall Game" {
		t.Fatalf("commands %v", got)
	}
	if d.runningExecutable != "" {
		t.Fatal("runningExecutable must be cleared")
	}
}

func TestStopAppFailure(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps}
	boom := errors.New("no process")
	e.runner.Script("killall Game", nil, boom)
	if err := d.StopApp(writeBundle(t)); !errors.Is(err, boom) {
		t.Fatalf("error %v, want boom", err)
	}
}

const (
	windowLine = `osascript -e tell application "Game" to id of window 1`
)

func TestScreenshotCapturesTheAppWindow(t *testing.T) {
	e := newTestEnv(t)
	if err := e.deps.ensureDirs(); err != nil {
		t.Fatal(err)
	}
	d := &Device{deps: e.deps, runningExecutable: "Game"}
	file := filepath.Join(e.deps.ScreenshotDir, "42.png")
	e.runner.Script(windowLine, []byte("42\n"), nil)
	e.runner.Script("screencapture -l 42 "+file, nil, nil)
	if err := os.WriteFile(file, []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}

	data, w, h, err := d.GetScreenshot()
	if err != nil || string(data) != "png" || w != 0 || h != 0 {
		t.Fatalf("got %q %d %d %v", data, w, h, err)
	}
	want := []string{windowLine, "screencapture -l 42 " + file}
	if got := e.runner.Commands(); len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("commands %v, want %v", got, want)
	}
}

func TestScreenshotErrors(t *testing.T) {
	file := func(e *testEnv) string { return filepath.Join(e.deps.ScreenshotDir, "42.png") }
	tests := []struct {
		name   string
		script func(e *testEnv)
	}{
		{"no window", func(*testEnv) {}},
		{"capture fails", func(e *testEnv) {
			e.runner.Script(windowLine, []byte("42"), nil)
		}},
		{"file missing", func(e *testEnv) {
			e.runner.Script(windowLine, []byte("42"), nil)
			e.runner.Script("screencapture -l 42 "+file(e), nil, nil)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEnv(t)
			tc.script(e)
			d := &Device{deps: e.deps, runningExecutable: "Game"}
			if _, _, _, err := d.GetScreenshot(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestHandlerUnsupportedOperations(t *testing.T) {
	h := NewHandlerWith(nil, newTestEnv(t).deps)
	if err := h.StartDevice("x"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StartDevice: %v", err)
	}
	if err := h.StopDevice("x"); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StopDevice: %v", err)
	}
	if _, err := h.RegisterDevice(device.RegisterData{}); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("RegisterDevice: %v", err)
	}
	if h.Name() != Manager || h.Start() != nil || h.Stop() != nil {
		t.Error("Name, Start and Stop")
	}
}
