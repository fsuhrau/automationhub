package androidsdk

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeEnv is a finder on a machine with the programs on PATH and the
// environment.
func fakeEnv(onPath map[string]string, env map[string]string, os string) Finder {
	return Finder{
		LookPath: func(name string) (string, error) {
			if p, ok := onPath[name]; ok {
				return p, nil
			}
			return "", exec.ErrNotFound
		},
		Getenv: func(key string) string { return env[key] },
		GOOS:   os,
	}
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
	f := fakeEnv(map[string]string{"aapt": "/usr/local/bin/aapt"}, map[string]string{"ANDROID_HOME": sdk}, "darwin")
	if p, err := f.AAPT(); err != nil || p != "/usr/local/bin/aapt" {
		t.Fatalf("%q %v", p, err)
	}
}

func TestNewestBuildToolsOfAndroidHome(t *testing.T) {
	sdk := t.TempDir()
	for _, v := range []string{"9.0.0", "30.0.3", "34.0.0", "35.0.0-rc1"} {
		tool(t, filepath.Join(sdk, "build-tools", v, "aapt"), true)
	}
	f := fakeEnv(nil, map[string]string{"ANDROID_HOME": sdk}, "darwin")
	if p, err := f.AAPT(); err != nil || p != filepath.Join(sdk, "build-tools", "35.0.0-rc1", "aapt") {
		t.Fatalf("%q %v", p, err)
	}
	// a release beats its release candidate
	tool(t, filepath.Join(sdk, "build-tools", "35.0.0", "aapt"), true)
	if p, _ := f.AAPT(); p != filepath.Join(sdk, "build-tools", "35.0.0", "aapt") {
		t.Fatalf("release: %q", p)
	}
}

func TestFallbacks(t *testing.T) {
	sdkRoot := t.TempDir()
	home := t.TempDir()
	// newest build-tools only has aapt2, an older one has a non-executable aapt
	tool(t, filepath.Join(sdkRoot, "build-tools", "36.0.0", "aapt2"), true)
	tool(t, filepath.Join(sdkRoot, "build-tools", "33.0.0", "aapt"), false)
	f := fakeEnv(nil, map[string]string{"ANDROID_HOME": filepath.Join(home, "missing"), "ANDROID_SDK_ROOT": sdkRoot, "HOME": home}, "linux")
	if p, err := f.AAPT(); err != nil || p != filepath.Join(sdkRoot, "build-tools", "36.0.0", "aapt2") {
		t.Fatalf("%q %v", p, err)
	}

	// the default macOS location
	tool(t, filepath.Join(home, "Library", "Android", "sdk", "build-tools", "34.0.0", "aapt"), true)
	f = fakeEnv(nil, map[string]string{"HOME": home}, "darwin")
	if p, err := f.AAPT(); err != nil || !strings.HasSuffix(p, filepath.Join("Library", "Android", "sdk", "build-tools", "34.0.0", "aapt")) {
		t.Fatalf("default sdk: %q %v", p, err)
	}
}

func TestNotFound(t *testing.T) {
	home := t.TempDir()
	f := fakeEnv(nil, map[string]string{"HOME": home, "ANDROID_HOME": "/nowhere"}, "darwin")
	_, err := f.AAPT()
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "/nowhere") || !strings.Contains(err.Error(), "ANDROID_HOME") {
		t.Fatalf("%v", err)
	}
}

func TestSystemFinderLooksOnThisMachine(t *testing.T) {
	f := SystemFinder()

	if f.LookPath == nil || f.Getenv == nil || f.GOOS == "" {
		t.Fatalf("finder %+v", f)
	}
	// the package function runs the same search; either outcome is valid on
	// a machine, but a path must name an existing file
	if p, err := AAPT(); err == nil {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Fatalf("%q: %v", p, statErr)
		}
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestWindowsDefaultLocation(t *testing.T) {
	local := t.TempDir()
	// Windows has no executable bit: any file counts
	tool(t, filepath.Join(local, "Android", "Sdk", "build-tools", "34.0.0", "aapt.exe"), false)
	tool(t, filepath.Join(local, "Android", "Sdk", "build-tools", "34.0.0", "aapt"), false)
	f := fakeEnv(nil, map[string]string{"HOME": "C:/Users/me", "LOCALAPPDATA": local}, "windows")

	p, err := f.AAPT()

	if err != nil || p != filepath.Join(local, "Android", "Sdk", "build-tools", "34.0.0", "aapt") {
		t.Fatalf("%q %v", p, err)
	}
}
