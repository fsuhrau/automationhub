package iosdevice

import (
	"errors"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
)

// goIOSRig is a rig for an iOS 16 device, which starts WebDriverAgent with go-ios.
func goIOSRig(t *testing.T) *rig {
	t.Helper()
	r := newRigWith(t, &config.WebDriver{Launcher: LauncherGoIOS}, "U")
	_ = r.h.RefreshDevices(true)
	return r
}

// logsURLWhenStarted makes the go-ios runner log the WebDriverAgent URL like
// the library does, then block like a running runner until the returned
// channel is closed.
func (r *rig) logsURLWhenStarted(url string) chan struct{} {
	stop := make(chan struct{})
	r.usb.xcuiFn = func() error {
		r.log.Info("ServerURLHere->" + url + "<-ServerURLHere")
		<-stop
		return nil
	}
	return stop
}

func TestGoIOSLauncherStartsTheRunnerAndReadsTheURLFromTheLog(t *testing.T) {
	r := goIOSRig(t)
	stop := r.logsURLWhenStarted(r.wda.URL)
	defer close(stop)
	if err := r.h.StartDevice("U"); err != nil {
		t.Fatal(err)
	}
	r.usb.mu.Lock()
	runs := r.usb.xcuiRuns
	r.usb.mu.Unlock()
	if len(runs) != 1 || runs[0] != "com.automationhub.WebDriverAgentRunner.xctrunner WebDriverAgentRunner.xctest" {
		t.Errorf("runs = %q", runs)
	}
	if len(r.proc.runner.Commands()) != 0 {
		t.Error("xcodebuild must not run for the go-ios launcher")
	}
	r.wda.saw(t, "POST /session")
	if err := r.h.StopDevice("U"); err != nil {
		t.Fatal(err)
	}
	if r.usb.xcuiClose != 1 {
		t.Errorf("CloseXCUI called %d times", r.usb.xcuiClose)
	}
}

func TestGoIOSLauncherReportsARunnerThatStops(t *testing.T) {
	r := goIOSRig(t)
	r.usb.xcuiErr = errors.New("developer image not mounted")
	err := r.h.StartDevice("U")
	if err == nil || !strings.Contains(err.Error(), "developer image not mounted") {
		t.Fatalf("err = %v", err)
	}
}

func TestGoIOSLauncherTimesOutWithoutAURL(t *testing.T) {
	r := goIOSRig(t)
	stop := make(chan struct{})
	defer close(stop)
	r.usb.xcuiFn = func() error { <-stop; return nil }
	errc := make(chan error, 1)
	go func() { errc <- r.h.StartDevice("U") }()
	waitPending(t, r.clock, 1)
	r.clock.Advance(WDAStartTimeout)
	if err := <-errc; err == nil || !strings.Contains(err.Error(), "did not report its server URL") {
		t.Fatalf("err = %v", err)
	}
}

func TestGoIOSLauncherNeedsTheDevice(t *testing.T) {
	r := goIOSRig(t)
	r.usb.serials = nil
	if err := r.h.StartDevice("U"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err = %v", err)
	}
}
