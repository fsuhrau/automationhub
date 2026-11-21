package macos

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/hostinfo/hostinfotest"
)

func TestDepsDefaults(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	d := Deps{}.withDefaults()
	if d.InstallDir != "/home/tester/.automationhub/installs" {
		t.Errorf("InstallDir %q", d.InstallDir)
	}
	if d.ScreenshotDir != filepath.Join(os.TempDir(), "automation_hub") {
		t.Errorf("ScreenshotDir %q", d.ScreenshotDir)
	}
	if d.Runner == nil || d.Clock == nil || d.GOOS == "" {
		t.Errorf("env defaults: %+v", d.Env)
	}
	if NewHandler(nil).deps.InstallDir != d.InstallDir {
		t.Error("NewHandler must use the defaults")
	}
}

func TestInitCreatesDirectories(t *testing.T) {
	e := newTestEnv(t)
	e.scriptMac()
	h := NewHandlerWith(hostinfotest.NewStore(nil), e.deps)
	if err := h.Init("master:1", "node", nil); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{e.deps.InstallDir, e.deps.ScreenshotDir} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("%s not created: %v", dir, err)
		}
	}
}

func TestInitFailures(t *testing.T) {
	boom := errors.New("db down")
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Run("directory", func(t *testing.T) {
		e := newTestEnv(t)
		e.deps.InstallDir = filepath.Join(blocker, "installs")
		if err := NewHandlerWith(hostinfotest.NewStore(nil), e.deps).Init("m", "n", nil); err == nil {
			t.Fatal("expected a directory error")
		}
	})
	t.Run("storage", func(t *testing.T) {
		e := newTestEnv(t)
		e.scriptMac()
		if err := NewHandlerWith(hostinfotest.NewStore(boom), e.deps).Init("m", "n", nil); !errors.Is(err, boom) {
			t.Fatalf("error %v, want boom", err)
		}
	})
}

// TestHandlerRefreshThenActions is the functional test: Init registers the Mac,
// then the device installs, starts, captures and stops an app.
func TestHandlerRefreshThenActions(t *testing.T) {
	e := newTestEnv(t)
	e.scriptMac()
	store := hostinfotest.NewStore(nil)
	var h device.Handler = NewHandlerWith(store, e.deps)

	if err := h.Init("master:1", "node", nil); err != nil {
		t.Fatal(err)
	}
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	devs, err := h.GetDevices()
	if err != nil || len(devs) != 1 || devs[0].DeviceID() != testUUID {
		t.Fatalf("devices %v, %v", devs, err)
	}
	if !h.HasDevice(devs[0]) || len(store.Updated()) != 1 {
		t.Fatalf("HasDevice / stored: %v", store.Updated())
	}
	runActions(t, e, devs[0])
}

func runActions(t *testing.T, e *testEnv, d device.Device) {
	t.Helper()
	params := writeBundle(t)
	if err := d.InstallApp(params); err != nil {
		t.Fatal(err)
	}
	e.runner.Script(startLine(e), nil, nil)
	e.runner.Script("killall Game", nil, nil)
	e.runner.Script(windowLine, []byte("7"), nil)
	file := filepath.Join(e.deps.ScreenshotDir, "7.png")
	e.runner.Script("screencapture -l 7 "+file, nil, nil)
	if err := os.WriteFile(file, []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := d.StartApp(nil, params, "s1", "n:1"); err != nil {
		t.Fatal(err)
	}
	if data, _, _, err := d.GetScreenshot(); err != nil || string(data) != "png" {
		t.Fatalf("screenshot %q, %v", data, err)
	}
	if err := d.StopApp(params); err != nil {
		t.Fatal(err)
	}
}
