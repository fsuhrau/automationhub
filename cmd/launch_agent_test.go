package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLaunchAgentRunsTheMaster(t *testing.T) {
	a := newLaunchAgent("/opt/bin", "/etc/hub.yaml")

	if a.Program != "/opt/bin/hub" || strings.Join(a.ProgramArguments, " ") != "master --config /etc/hub.yaml" || !a.KeepAlive || !a.RunAtLoad {
		t.Errorf("agent = %+v", a)
	}
}

func TestLaunchAgentPath(t *testing.T) {
	if got, want := launchAgentPath("/home/u"), "/home/u/Library/LaunchAgents/com.automationhub.master.plist"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteLaunchAgentRendersThePlist(t *testing.T) {
	path := launchAgentPath(t.TempDir())

	if err := writeLaunchAgent(path, newLaunchAgent("/opt/bin", "/etc/hub.yaml")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<string>com.automationhub.master</string>", "<string>/opt/bin/hub</string>", "<string>/etc/hub.yaml</string>", "<key>KeepAlive</key><true/>"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("plist lacks %s:\n%s", want, data)
		}
	}
}

func TestWriteLaunchAgentFailures(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	tests := map[string]struct{ path, want string }{
		"directory is a file": {filepath.Join(file, "x.plist"), "create LaunchAgents directory"},
		"path is a directory": {dir, "create plist"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := writeLaunchAgent(tc.path, newLaunchAgent("/b", "/c"))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestWriteLaunchAgentFailsOnAFullDisk(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("no /dev/full")
	}
	// a symlink to /dev/full makes the template write fail
	dir := t.TempDir()
	path := filepath.Join(dir, "a.plist")
	if err := os.Symlink("/dev/full", path); err != nil {
		t.Fatal(err)
	}

	err := writeLaunchAgent(path, newLaunchAgent("/b", "/c"))

	if err == nil || !strings.Contains(err.Error(), "render plist") {
		t.Errorf("err = %v", err)
	}
}
