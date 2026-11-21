package iosdevice

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
)

func TestUseXcodebuild(t *testing.T) {
	for _, tc := range []struct {
		launcher, version string
		want              bool
	}{
		{"", "16.7.2", false},
		{"", "17.0", true},
		{"", "18.7.11", true},
		{"auto", "27.0.1", true},
		{"go-ios", "18.7", false},
		{"xcodebuild", "15.1", true},
		{"", "", false},
	} {
		if got := useXcodebuild(&config.WebDriver{Launcher: tc.launcher}, tc.version); got != tc.want {
			t.Errorf("useXcodebuild(%q, %q) = %v", tc.launcher, tc.version, got)
		}
	}
	if !useXcodebuild(nil, "18.0") {
		t.Error("no config on iOS 18 should use xcodebuild")
	}
}

func TestXcodebuildArgs(t *testing.T) {
	if _, err := xcodebuildArgs(nil, "u"); err == nil || !strings.Contains(err.Error(), "webdriver.xctestrun") {
		t.Fatalf("missing config: %v", err)
	}
	got, _ := xcodebuildArgs(&config.WebDriver{XCTestRun: "/wda/WDA.xctestrun", Project: "ignored"}, "UDID")
	want := []string{"test-without-building", "-xctestrun", "/wda/WDA.xctestrun", "-destination", "id=UDID"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("xctestrun args = %v", got)
	}
	got, _ = xcodebuildArgs(&config.WebDriver{Project: "/wda/WebDriverAgent.xcodeproj", DerivedData: "/tmp/dd"}, "UDID")
	want = []string{"test", "-project", "/wda/WebDriverAgent.xcodeproj", "-scheme", "WebDriverAgentRunner", "-destination", "id=UDID", "-allowProvisioningUpdates", "-derivedDataPath", "/tmp/dd"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("project args = %v", got)
	}
	got, _ = xcodebuildArgs(&config.WebDriver{Project: "p.xcodeproj", Scheme: "Custom"}, "U")
	if got[3] != "-scheme" || got[4] != "Custom" {
		t.Errorf("scheme args = %v", got)
	}
}

func TestLineTailKeepsTheLastLines(t *testing.T) {
	tail := newLineTail(2)
	for _, l := range []string{"a", "b", "c"} {
		tail.add(l)
	}
	if got := tail.String(); got != "b\nc" {
		t.Errorf("tail = %q", got)
	}
}

// killOnly is a process that can't be interrupted, only killed.
type killOnly struct{ killed int }

func (p *killOnly) Wait() error { return nil }
func (p *killOnly) Kill() error { p.killed++; return nil }

func TestStopKillsAProcessThatCannotBeInterrupted(t *testing.T) {
	proc := &killOnly{}
	done := make(chan struct{})
	close(done)
	r := &wdaRunner{proc: proc, done: done, after: func(time.Duration) <-chan time.Time { return nil }}
	r.Stop()
	if proc.killed != 1 {
		t.Fatalf("killed %d times, want once", proc.killed)
	}
}
