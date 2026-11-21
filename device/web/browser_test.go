package web

import (
	"testing"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		goos, browser, path, name string
		args                      []string
	}{
		{"darwin", "safari", safariApp, "defaults", []string{"read", safariApp + "/Contents/Info", "CFBundleShortVersionString"}},
		{"darwin", "chrome", chromeBin, chromeBin, []string{"--version"}},
		{"linux", "firefox", "/usr/bin/firefox", "/usr/bin/firefox", []string{"--version"}},
		{"windows", "safari", "C:/safari.exe", "C:/safari.exe", []string{"--version"}},
	}
	for _, tc := range tests {
		name, args := versionCommand(tc.goos, tc.browser, tc.path)
		if name != tc.name || len(args) != len(tc.args) || args[0] != tc.args[0] || args[len(args)-1] != tc.args[len(tc.args)-1] {
			t.Errorf("%s/%s: got %s %v, want %s %v", tc.goos, tc.browser, name, args, tc.name, tc.args)
		}
	}
}

func TestParseBrowserVersion(t *testing.T) {
	tests := map[string]string{
		"Google Chrome 126.0.1\n": "126.0.1",
		"Mozilla Firefox 127.0":   "127.0",
		"17.5\n":                  "17.5",
		"":                        "",
	}
	for in, want := range tests {
		if got := parseBrowserVersion(in); got != want {
			t.Errorf("parseBrowserVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisplayName(t *testing.T) {
	if got := displayName("chrome", "126.0", "mac-1"); got != "chrome (126.0) @ mac-1" {
		t.Errorf("with version: %q", got)
	}
	if got := displayName("chrome", "", "mac-1"); got != "chrome @ mac-1" {
		t.Errorf("without version: %q", got)
	}
}
