package macos

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	testUUID  = "54DECB62-3993-4031-9C6A-18CE048CC63C"
	ioregLine = "ioreg -rd1 -c IOPlatformExpertDevice"
	hardware  = "      Model Identifier: MacBookPro18,1\n      Serial Number (system): C02XYZ\n"
	displays  = "      Chipset Model: Apple M1 Pro\n      Metal Support: Metal 3\n"
)

// testEnv is a Deps with directories under t.TempDir and a scripted Mac.
type testEnv struct {
	deps   Deps
	runner *testkit.FakeRunner
	clock  *testkit.FakeClock
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	runner := &testkit.FakeRunner{}
	clock := testkit.NewFakeClock(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	root := t.TempDir()
	deps := Deps{
		Env:           hostinfo.Env{Runner: runner, Clock: clock, GOOS: "darwin"},
		InstallDir:    filepath.Join(root, "installs"),
		ScreenshotDir: filepath.Join(root, "shots"),
	}
	return &testEnv{deps: deps.withDefaults(), runner: runner, clock: clock}
}

// scriptMac scripts every command that reading the Mac's infos runs.
func (e *testEnv) scriptMac() {
	out := func(line, text string) { e.runner.Script(line, []byte(text), nil) }
	out("sw_vers -productName", "macOS\n")
	out("sw_vers -productVersion", "14.5\n")
	out("sysctl -n kern.hostname", "mac-1\n")
	out(ioregLine, "    \"IOPlatformUUID\" = \""+testUUID+"\"\n")
	out("system_profiler SPHardwareDataType", hardware)
	out("system_profiler SPDisplaysDataType", displays)
	out("sysctl -n machdep.cpu.brand_string", "Apple M1 Pro\n")
	out("sysctl -n hw.memsize", "17179869184\n")
}

// writeBundle writes a zip holding the app executable and returns its parameters.
func writeBundle(t *testing.T) *app.Parameter {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bundle.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("Game.app/Contents/MacOS/Game")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("#!/bin/sh\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return &app.Parameter{
		Name: "Game.app",
		App: &app.AppParams{
			AppPath:    path,
			Hash:       "h1",
			Executable: &app.ExecutableParams{Executable: "Contents/MacOS/Game"},
		},
	}
}

type appExe = app.ExecutableParams

// paramsWith returns a copy of params that uses the given app parameters.
func paramsWith(params *app.Parameter, a *app.AppParams) *app.Parameter {
	p := *params
	p.App = a
	return &p
}
