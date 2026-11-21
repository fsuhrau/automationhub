package macos

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestNewDeviceReadsTheMac(t *testing.T) {
	e := newTestEnv(t)
	e.scriptMac()
	d := newDevice(context.Background(), e.deps)

	if d.DeviceID() != testUUID || d.DeviceName() != "mac-1" || d.DeviceOSName() != "macOS" || d.DeviceOSVersion() != "14.5" {
		t.Fatalf("identity: %q %q %q %q", d.DeviceID(), d.DeviceName(), d.DeviceOSName(), d.DeviceOSVersion())
	}
	if d.DeviceState() != device.StateBooted || d.DeviceType() != int(models.DeviceTypeDesktop) || d.PlatformType() != int(models.PlatformTypeMac) || d.TargetVersion() != "" {
		t.Fatalf("kind: %v %d %d %q", d.DeviceState(), d.DeviceType(), d.PlatformType(), d.TargetVersion())
	}
	want := map[string]string{
		hostinfo.ParamModel: "MacBookPro18,1", hostinfo.ParamSerial: "C02XYZ", hostinfo.ParamCPU: "Apple M1 Pro",
		hostinfo.ParamGPU: "Apple M1 Pro", hostinfo.ParamRAM: "16 GB", hostinfo.ParamDriver: "Metal 3",
	}
	for k, v := range want {
		if got := d.DeviceParameter()[k]; got != v {
			t.Errorf("parameter %s = %q, want %q", k, got, v)
		}
	}
	if !d.UpdatedAt().Equal(e.clock.Now()) {
		t.Errorf("updatedAt %v, want clock time", d.UpdatedAt())
	}
}

func TestDeviceTrivia(t *testing.T) {
	d := &Device{}
	if d.IsAppConnected() || d.ConnectionTimeout() != ConnectionTimeout {
		t.Fatal("a new device is not connected")
	}
}

func TestRoles(t *testing.T) {
	if got, want := testkit.Roles(&Device{}), []string{"Screenshotter"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
}

func TestInstallLifecycle(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps}
	params := writeBundle(t)

	if ok, err := d.IsAppInstalled(params); ok || err != nil {
		t.Fatalf("before install: %v, %v", ok, err)
	}
	if err := d.InstallApp(params); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(e.deps.InstallDir, "h1", "Game.app", "Contents/MacOS/Game")
	info, err := os.Stat(exe)
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Fatalf("executable: %v, %v", info, err)
	}
	if ok, err := d.IsAppInstalled(params); !ok || err != nil {
		t.Fatalf("after install: %v, %v", ok, err)
	}
	if err := d.UninstallApp(params); err != nil {
		t.Fatal(err)
	}
	if ok, _ := d.IsAppInstalled(params); ok {
		t.Fatal("still installed after uninstall")
	}
}

func TestInstallErrors(t *testing.T) {
	e := newTestEnv(t)
	d := &Device{deps: e.deps}
	params := writeBundle(t)

	missing := *params.App
	missing.AppPath = filepath.Join(t.TempDir(), "absent.zip")
	if err := d.InstallApp(paramsWith(params, &missing)); err == nil {
		t.Error("unzip of a missing archive must fail")
	}
	wrongExe := *params.App
	wrongExe.Executable = &app.ExecutableParams{Executable: "Contents/MacOS/Other"}
	if err := d.InstallApp(paramsWith(params, &wrongExe)); err == nil {
		t.Error("chmod of a missing executable must fail")
	}
}

func TestIsAppInstalledStatError(t *testing.T) {
	e := newTestEnv(t)
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	e.deps.InstallDir = filepath.Join(file, "sub") // the parent is a file: stat fails with ENOTDIR
	d := &Device{deps: e.deps}
	if _, err := d.IsAppInstalled(writeBundle(t)); err == nil {
		t.Fatal("expected a stat error")
	}
}

func TestUninstallError(t *testing.T) {
	e := newTestEnv(t)
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	e.deps.InstallDir = filepath.Join(file, "sub")
	d := &Device{deps: e.deps}
	if err := d.UninstallApp(writeBundle(t)); err == nil {
		t.Fatal("expected a remove error")
	}
}
