package androidsdk

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeEnv replaces PATH lookup and the environment.
func fakeEnv(t *testing.T, onPath map[string]string, env map[string]string, os string) {
	t.Helper()
	oldLook, oldEnv, oldOS := lookPath, getenv, goos
	t.Cleanup(func() { lookPath, getenv, goos = oldLook, oldEnv, oldOS })
	lookPath = func(name string) (string, error) {
		if p, ok := onPath[name]; ok {
			return p, nil
		}
		return "", exec.ErrNotFound
	}
	getenv = func(key string) string { return env[key] }
	goos = os
}

func tool(t *testing.T, path string, executable bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	mode := os.FileMode(0o644)
	if executable {
		mode = 0o755
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}

func TestPathWins(t *testing.T) {
	sdk := t.TempDir()
	tool(t, filepath.Join(sdk, "build-tools", "34.0.0", "aapt"), true)
	fakeEnv(t, map[string]string{"aapt": "/usr/local/bin/aapt"}, map[string]string{"ANDROID_HOME": sdk}, "darwin")
	if p, err := AAPT(); err != nil || p != "/usr/local/bin/aapt" {
		t.Fatalf("%q %v", p, err)
	}
}

func TestNewestBuildToolsOfAndroidHome(t *testing.T) {
	sdk := t.TempDir()
	for _, v := range []string{"9.0.0", "30.0.3", "34.0.0", "35.0.0-rc1"} {
		tool(t, filepath.Join(sdk, "build-tools", v, "aapt"), true)
	}
	fakeEnv(t, nil, map[string]string{"ANDROID_HOME": sdk}, "darwin")
	if p, err := AAPT(); err != nil || p != filepath.Join(sdk, "build-tools", "35.0.0-rc1", "aapt") {
		t.Fatalf("%q %v", p, err)
	}
	// a release beats its release candidate
	tool(t, filepath.Join(sdk, "build-tools", "35.0.0", "aapt"), true)
	if p, _ := AAPT(); p != filepath.Join(sdk, "build-tools", "35.0.0", "aapt") {
		t.Fatalf("release: %q", p)
	}
}

func TestFallbacks(t *testing.T) {
	sdkRoot := t.TempDir()
	home := t.TempDir()
	// newest build-tools only has aapt2, an older one has a non-executable aapt
	tool(t, filepath.Join(sdkRoot, "build-tools", "36.0.0", "aapt2"), true)
	tool(t, filepath.Join(sdkRoot, "build-tools", "33.0.0", "aapt"), false)
	fakeEnv(t, nil, map[string]string{"ANDROID_HOME": filepath.Join(home, "missing"), "ANDROID_SDK_ROOT": sdkRoot, "HOME": home}, "linux")
	if p, err := AAPT(); err != nil || p != filepath.Join(sdkRoot, "build-tools", "36.0.0", "aapt2") {
		t.Fatalf("%q %v", p, err)
	}

	// the default macOS location
	tool(t, filepath.Join(home, "Library", "Android", "sdk", "build-tools", "34.0.0", "aapt"), true)
	fakeEnv(t, nil, map[string]string{"HOME": home}, "darwin")
	if p, err := AAPT(); err != nil || !strings.HasSuffix(p, filepath.Join("Library", "Android", "sdk", "build-tools", "34.0.0", "aapt")) {
		t.Fatalf("default sdk: %q %v", p, err)
	}
}

func TestNotFound(t *testing.T) {
	home := t.TempDir()
	fakeEnv(t, nil, map[string]string{"HOME": home, "ANDROID_HOME": "/nowhere"}, "darwin")
	_, err := AAPT()
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "/nowhere") || !strings.Contains(err.Error(), "ANDROID_HOME") {
		t.Fatalf("%v", err)
	}
}
