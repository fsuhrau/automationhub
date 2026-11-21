package iosdevice

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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

// fakeXcodebuild puts a script named xcodebuild first on PATH.
func fakeXcodebuild(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "xcodebuild"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestExecStarterStreamsOutputAndStopsTheProcessGroup(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	fakeXcodebuild(t, "echo $$ > "+pidFile+"\necho started\nexec sleep 60\n")
	var out syncBuffer
	proc, err := execStarter{}.StartStreaming(&out, "xcodebuild", "test")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return strings.Contains(out.String(), "started") })
	if err := proc.Interrupt(); err != nil {
		t.Fatal(err)
	}
	_ = proc.Wait()
}

func TestExecStarterKill(t *testing.T) {
	fakeXcodebuild(t, "trap '' INT\nexec sleep 60\n")
	proc, err := execStarter{}.StartStreaming(&syncBuffer{}, "xcodebuild")
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = proc.Wait()
}

func TestExecStarterReportsAMissingCommand(t *testing.T) {
	if _, err := (execStarter{}).StartStreaming(&syncBuffer{}, "/no/such/xcodebuild"); err == nil || !strings.Contains(err.Error(), "start /no/such/xcodebuild") {
		t.Fatalf("err = %v", err)
	}
}
