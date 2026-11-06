package iosdevice

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/modules/webdriver"
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
}

func TestParseServerURL(t *testing.T) {
	line := "2026-11-06 10:00:00.000 WebDriverAgentRunner-Runner[123:456] ServerURLHere->http://192.168.1.20:8100<-ServerURLHere"
	if got, ok := webdriver.ParseServerURL(line); !ok || got != "http://192.168.1.20:8100" {
		t.Fatalf("ParseServerURL = %q, %v", got, ok)
	}
	if _, ok := webdriver.ParseServerURL("Test Suite 'All tests' started"); ok {
		t.Fatal("unrelated line parsed")
	}
}

// fakeXcodebuild puts a script named xcodebuild first on PATH.
func fakeXcodebuild(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "xcodebuild"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestXcodebuildRunnerReportsTheOutputWhenItFails(t *testing.T) {
	fakeXcodebuild(t, "echo 'Testing failed:'\necho 'Developer Mode disabled' >&2\nexit 65\n")
	_, err := startXcodebuildRunner(&config.WebDriver{XCTestRun: "x.xctestrun"}, "UDID", t.Logf)
	if err == nil || !strings.Contains(err.Error(), "Developer Mode disabled") || !strings.Contains(err.Error(), "Testing failed:") {
		t.Fatalf("err = %v", err)
	}
}

func TestXcodebuildRunnerStopsXcodebuildWhenForwardingFails(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	fakeXcodebuild(t, "echo $$ > "+pidFile+"\necho 'ServerURLHere->http://10.0.0.2:8100<-ServerURLHere'\nexec sleep 60\n")
	// there is no such device (and no usbmuxd here), so the forward fails
	if _, err := startXcodebuildRunner(&config.WebDriver{XCTestRun: "x.xctestrun"}, "no-such-device", t.Logf); err == nil {
		t.Skip("a device with this id exists")
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("/proc/" + strings.TrimSpace(string(data))); err == nil {
		t.Fatal("xcodebuild still runs")
	}
}
