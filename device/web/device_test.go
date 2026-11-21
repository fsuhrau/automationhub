package web

import (
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestNewDeviceOnMac(t *testing.T) {
	e := newTestEnv("darwin")
	e.scriptMac()
	d := newDevice(t.Context(), e.env, browserConfig{"chrome", chromeBin})

	if d.DeviceID() != uuid+"/chrome" || d.DeviceName() != "chrome (126.0.1) @ mac-1" || d.TargetVersion() != "126.0.1" {
		t.Fatalf("identity: %q %q %q", d.DeviceID(), d.DeviceName(), d.TargetVersion())
	}
	if d.DeviceOSName() != "macOS" || d.DeviceOSVersion() != "14.5" || d.DeviceState() != device.StateBooted {
		t.Fatalf("os: %q %q %v", d.DeviceOSName(), d.DeviceOSVersion(), d.DeviceState())
	}
	if d.DeviceType() != int(models.DeviceTypeBrowser) || d.PlatformType() != int(models.PlatformTypeWeb) {
		t.Fatalf("kind: %d %d", d.DeviceType(), d.PlatformType())
	}
	p := d.DeviceParameter()
	if p[hostinfo.ParamModel] != "Mac14,2" || p[hostinfo.ParamSerial] != "C02XYZ" || p["Browser Version"] != "126.0.1" {
		t.Fatalf("parameters %v", p)
	}
	if !d.UpdatedAt().Equal(e.env.Clock.Now()) {
		t.Errorf("updatedAt %v", d.UpdatedAt())
	}
}

func TestNewDeviceOnLinuxWithoutVersion(t *testing.T) {
	e := newTestEnv("linux")
	e.runner.Script("hostname", []byte("box\n"), nil)
	d := newDevice(t.Context(), e.env.WithDefaults(), browserConfig{"firefox", "/usr/bin/firefox"})
	if d.DeviceName() != "firefox @ box" || d.DeviceOSName() != "linux" || d.TargetVersion() != "" {
		t.Fatalf("got %q %q %q", d.DeviceName(), d.DeviceOSName(), d.TargetVersion())
	}
	if _, ok := d.DeviceParameter()["Browser Version"]; ok {
		t.Fatal("no version parameter without a version")
	}
}

func TestDeviceTrivia(t *testing.T) {
	d := &Device{}
	if d.IsAppConnected() || d.ConnectionTimeout() != ConnectionTimeout {
		t.Fatal("a new device is not connected")
	}
	if ok, err := d.IsAppInstalled(nil); !ok || err != nil {
		t.Fatalf("IsAppInstalled %v %v", ok, err)
	}
	if d.InstallApp(nil) != nil || d.UninstallApp(nil) != nil {
		t.Fatal("install and uninstall have nothing to do")
	}
}

func TestHasNoOptionalRoles(t *testing.T) {
	if got := testkit.Roles(&Device{}); len(got) != 0 {
		t.Fatalf("roles = %v, want none", got)
	}
}
