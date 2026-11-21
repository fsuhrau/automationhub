package web

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/hostinfo/hostinfotest"
)

func TestInitFailures(t *testing.T) {
	e := newTestEnv("darwin")
	e.scriptMac()
	cfg := config.Manager{Browser: map[string]string{"chrome": chromeBin}}
	boom := errors.New("db down")
	if err := NewHandlerWith(cfg, hostinfotest.NewStore(boom), e.env).Init("m", "n", nil); !errors.Is(err, boom) {
		t.Fatalf("error %v, want boom", err)
	}
}

func TestHandlerUnsupportedOperations(t *testing.T) {
	h := NewHandlerWith(config.Manager{}, nil, newTestEnv("linux").env)
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
	if NewHandler(config.Manager{}, nil).env.Runner == nil {
		t.Error("NewHandler must use the default environment")
	}
}

// TestHandlerRefreshThenActions is the functional test: Init registers a
// device per browser, then each device starts and stops its page.
func TestHandlerRefreshThenActions(t *testing.T) {
	e := newTestEnv("darwin")
	e.scriptMac()
	cfg := config.Manager{Browser: map[string]string{"chrome": chromeBin, "safari": safariApp}}
	store := hostinfotest.NewStore(nil)
	var h device.Handler = NewHandlerWith(cfg, store, e.env)

	if err := h.Init("master:1", "node", nil); err != nil {
		t.Fatal(err)
	}
	if err := h.RefreshDevices(true); err != nil {
		t.Fatal(err)
	}
	devs, err := h.GetDevices()
	if err != nil || len(devs) != 2 {
		t.Fatalf("devices %v, %v", devs, err)
	}
	if devs[0].DeviceID() != uuid+"/chrome" || devs[1].DeviceID() != uuid+"/safari" || devs[1].TargetVersion() != "17.5" {
		t.Fatalf("ids %q %q, safari version %q", devs[0].DeviceID(), devs[1].DeviceID(), devs[1].TargetVersion())
	}
	if !h.HasDevice(devs[0]) || len(store.Updated()) != 2 {
		t.Fatalf("HasDevice / stored %d", len(store.Updated()))
	}
	line := safariApp + " " + pageURL + "?sessionId=s1&nodeURL=n:1&deviceId=" + uuid + "/safari"
	e.runner.Script(line, nil, nil)
	if err := devs[1].StartApp(nil, webParams(), "s1", "n:1"); err != nil {
		t.Fatal(err)
	}
	if err := devs[1].StopApp(nil); err != nil {
		t.Fatal(err)
	}
}
