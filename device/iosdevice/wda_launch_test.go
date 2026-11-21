package iosdevice

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
)

func (r *rig) launcher(t *testing.T) *launcher {
	t.Helper()
	return &launcher{cfg: r.h.wdaConfig, deps: r.h.deps, logf: t.Logf}
}

func TestStartXcodebuildRunsTheExactCommandAndForwardsThePorts(t *testing.T) {
	r := newRig(t, "UDID1")
	runner, err := r.launcher(t).startXcodebuild("UDID1")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"xcodebuild test-without-building -xctestrun " + xctestrun + " -destination id=UDID1"}
	if got := r.proc.runner.Commands(); !reflect.DeepEqual(got, want) {
		t.Errorf("commands = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(r.usb.forwarded, []uint16{8100, 9100}) {
		t.Errorf("forwarded ports = %v", r.usb.forwarded)
	}
	if runner.Address != "http://"+r.wda.hostPort() || runner.MjpegAddress != "http://127.0.0.1:9100/" {
		t.Errorf("addresses = %s, %s", runner.Address, runner.MjpegAddress)
	}
	runner.Stop()
	if r.usb.closed != 2 || !reflect.DeepEqual(r.proc.proc.got(), []string{"interrupt"}) {
		t.Errorf("Stop closed %d forwards, signals %v", r.usb.closed, r.proc.proc.got())
	}
}

func TestStartXcodebuildBuildsTheProject(t *testing.T) {
	cfg := &config.WebDriver{Project: "/wda/WebDriverAgent.xcodeproj", DerivedData: "/dd"}
	r := newRigWith(t, cfg, "U")
	r.proc.runner.Script("xcodebuild test -project /wda/WebDriverAgent.xcodeproj -scheme WebDriverAgentRunner -destination id=U -allowProvisioningUpdates -derivedDataPath /dd", nil, nil)
	runner, err := r.launcher(t).startXcodebuild("U")
	if err != nil {
		t.Fatal(err)
	}
	runner.Stop()
}

func TestStartXcodebuildUsesThePortWebDriverAgentReported(t *testing.T) {
	r := newRig(t, "U")
	r.proc.output = "ServerURLHere->http://10.0.0.2:8200<-ServerURLHere\n"
	runner, err := r.launcher(t).startXcodebuild("U")
	if err != nil {
		t.Fatal(err)
	}
	runner.Stop()
	if !reflect.DeepEqual(r.usb.forwarded, []uint16{8200, 9100}) {
		t.Errorf("forwarded ports = %v", r.usb.forwarded)
	}
}

func TestDevicePortOf(t *testing.T) {
	for address, want := range map[string]int{
		"http://10.0.0.2:8300": 8300,
		"http://10.0.0.2":      wdaDevicePort,
		"://bad":               wdaDevicePort,
		"http://h:99999999999": wdaDevicePort,
	} {
		if got := devicePortOf(address); got != want {
			t.Errorf("devicePortOf(%q) = %d, want %d", address, got, want)
		}
	}
}

func TestStartXcodebuildErrors(t *testing.T) {
	t.Run("no config", func(t *testing.T) {
		r := newRigWith(t, nil, "U")
		if _, err := r.launcher(t).startXcodebuild("U"); err == nil || !strings.Contains(err.Error(), "webdriver.xctestrun") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("xcodebuild does not start", func(t *testing.T) {
		r := newRig(t, "U")
		if _, err := r.launcher(t).startXcodebuild("OTHER"); err == nil || !strings.Contains(err.Error(), "start xcodebuild") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("xcodebuild ends early", func(t *testing.T) {
		r := newRig(t, "U")
		r.proc.output = "Testing failed:\nDeveloper Mode disabled\n"
		r.proc.exits = true
		_, err := r.launcher(t).startXcodebuild("U")
		if err == nil || !strings.Contains(err.Error(), "Developer Mode disabled") || !strings.Contains(err.Error(), "Testing failed:") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("no device", func(t *testing.T) {
		r := newRig(t, "U")
		r.proc.runner.Script("xcodebuild test-without-building -xctestrun "+xctestrun+" -destination id=GONE", nil, nil)
		if _, err := r.launcher(t).startXcodebuild("GONE"); err == nil || !strings.Contains(err.Error(), "not found") {
			t.Fatalf("err = %v", err)
		}
		waitFor(t, func() bool { return len(r.proc.proc.got()) == 1 })
	})
	t.Run("forward fails and stops xcodebuild", func(t *testing.T) {
		r := newRig(t, "U")
		boom := errors.New("usbmux gone")
		ops := r.usb.ops()
		ops.Forward = func(deviceEntry, uint16) (Forwarder, error) { return nil, boom }
		r.h.deps.USB = ops
		if _, err := r.launcher(t).startXcodebuild("U"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
		waitFor(t, func() bool { return len(r.proc.proc.got()) == 1 })
	})
}

func TestStartXcodebuildTimesOutWithTheOutputSoFar(t *testing.T) {
	r := newRig(t, "U")
	r.proc.output = "Building…\n"
	errc := make(chan error, 1)
	go func() {
		_, err := r.launcher(t).startXcodebuild("U")
		errc <- err
	}()
	waitPending(t, r.clock, 1)
	r.clock.Advance(XcodebuildStartTimeout)
	err := <-errc
	if err == nil || !strings.Contains(err.Error(), "did not report its server URL") || !strings.Contains(err.Error(), "Building…") {
		t.Fatalf("err = %v", err)
	}
}

func TestStopKillsAnXcodebuildThatIgnoresTheInterrupt(t *testing.T) {
	r := newRig(t, "U")
	r.proc.stubborn = true
	runner, err := r.launcher(t).startXcodebuild("U")
	if err != nil {
		t.Fatal(err)
	}
	stopped := make(chan struct{})
	pending := r.clock.Pending() // the start timeout is still registered
	go func() { runner.Stop(); close(stopped) }()
	waitPending(t, r.clock, pending+1)
	r.clock.Advance(stopGracePeriod)
	<-stopped
	if got := r.proc.proc.got(); !reflect.DeepEqual(got, []string{"interrupt", "kill"}) {
		t.Errorf("signals = %v", got)
	}
}

func TestAttachUsesARunningWebDriverAgent(t *testing.T) {
	r := newRig(t, "U")
	r.usb.probeErr = nil
	runner, err := r.launcher(t).attach("U")
	if err != nil {
		t.Fatal(err)
	}
	if runner.Address != "http://"+r.wda.hostPort() || len(r.proc.runner.Commands()) != 0 {
		t.Errorf("address %s, commands %v", runner.Address, r.proc.runner.Commands())
	}
	runner.Stop() // only closes the forwards
	if r.usb.closed != 2 {
		t.Errorf("closed %d forwards", r.usb.closed)
	}
}

func TestAttachErrors(t *testing.T) {
	t.Run("no usbmux device", func(t *testing.T) {
		r := newRig(t)
		if _, err := r.launcher(t).attach("U"); err == nil || !strings.HasPrefix(err.Error(), "usbmux: ") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("nothing listens", func(t *testing.T) {
		r := newRig(t, "U")
		if _, err := r.launcher(t).attach("U"); err == nil || err.Error() != "nothing listens" {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("forward fails", func(t *testing.T) {
		r := newRig(t, "U")
		r.usb.probeErr = nil
		ops := r.usb.ops()
		ops.Forward = func(deviceEntry, uint16) (Forwarder, error) { return nil, errors.New("no port") }
		r.h.deps.USB = ops
		if _, err := r.launcher(t).attach("U"); err == nil || err.Error() != "no port" {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("does not answer like WebDriverAgent", func(t *testing.T) {
		r := newRig(t, "U")
		r.usb.probeErr = nil
		r.wda.Close()
		_, err := r.launcher(t).attach("U")
		if err == nil || !strings.Contains(err.Error(), "doesn't answer like WebDriverAgent") {
			t.Fatalf("err = %v", err)
		}
		if r.usb.closed != 2 {
			t.Errorf("closed %d forwards", r.usb.closed)
		}
	})
}

func TestXcodebuildOutputKeepsTheFirstURLAndTheLastLines(t *testing.T) {
	out := &xcodebuildOutput{found: make(chan string, 1), tail: newLineTail(2), scanned: make(chan struct{})}
	out.scan(strings.NewReader("ServerURLHere->http://a:1<-ServerURLHere\nServerURLHere->http://b:2<-ServerURLHere\nlast\n"))
	<-out.scanned
	if got := <-out.found; got != "http://a:1" {
		t.Errorf("found = %q", got)
	}
	if got := out.tail.String(); !strings.HasSuffix(got, "last") || strings.Contains(got, "a:1") {
		t.Errorf("tail = %q", got)
	}
}
