// Package androidsdk locates Android SDK tools for processes whose PATH lacks
// them, e.g. a master started outside of a shell with the SDK set up.
package androidsdk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// ErrNotFound means no aapt was found on PATH or in an SDK.
var ErrNotFound = errors.New("aapt not found: put the Android SDK build-tools on PATH or set ANDROID_HOME")

// lookPath and environment are replaced in tests.
var (
	lookPath = exec.LookPath
	getenv   = os.Getenv
	goos     = runtime.GOOS
)

// AAPT returns the aapt to run for "dump badging": aapt from PATH, else from
// the newest build-tools of $ANDROID_HOME, $ANDROID_SDK_ROOT or the default
// SDK location. aapt2 is used where aapt is missing; its "dump badging"
// prints the same format.
func AAPT() (string, error) {
	for _, name := range []string{"aapt", "aapt2"} {
		if p, err := lookPath(name); err == nil {
			return p, nil
		}
	}
	var searched []string
	for _, sdk := range sdkRoots() {
		for _, dir := range buildTools(sdk) {
			for _, name := range []string{"aapt", "aapt2"} {
				p := filepath.Join(dir, name)
				if isExecutable(p) {
					return p, nil
				}
			}
		}
		searched = append(searched, sdk)
	}
	if len(searched) == 0 {
		return "", ErrNotFound
	}
	return "", fmt.Errorf("%w (searched %s)", ErrNotFound, strings.Join(searched, ", "))
}

// sdkRoots lists the SDK folders to search, in order, without duplicates.
func sdkRoots() []string {
	var roots []string
	add := func(dir string) {
		if dir == "" {
			return
		}
		for _, r := range roots {
			if r == dir {
				return
			}
		}
		roots = append(roots, dir)
	}
	add(getenv("ANDROID_HOME"))
	add(getenv("ANDROID_SDK_ROOT"))
	if home := getenv("HOME"); home != "" {
		switch goos {
		case "darwin":
			add(filepath.Join(home, "Library", "Android", "sdk"))
		case "windows":
			add(filepath.Join(getenv("LOCALAPPDATA"), "Android", "Sdk"))
		default:
			add(filepath.Join(home, "Android", "Sdk"))
		}
	}
	return roots
}

// buildTools returns the build-tools folders of an SDK, newest first.
func buildTools(sdk string) []string {
	entries, err := os.ReadDir(filepath.Join(sdk, "build-tools"))
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool { return newer(versions[i], versions[j]) })
	dirs := make([]string, len(versions))
	for i, v := range versions {
		dirs[i] = filepath.Join(sdk, "build-tools", v)
	}
	return dirs
}

// newer compares build-tools versions like "34.0.0" and "35.0.0-rc1"; a
// release is newer than its release candidates.
func newer(a, b string) bool {
	an, apre := parseVersion(a)
	bn, bpre := parseVersion(b)
	for i := 0; i < len(an) || i < len(bn); i++ {
		var x, y int
		if i < len(an) {
			x = an[i]
		}
		if i < len(bn) {
			y = bn[i]
		}
		if x != y {
			return x > y
		}
	}
	if (apre == "") != (bpre == "") {
		return apre == ""
	}
	return apre > bpre
}

func parseVersion(v string) ([]int, string) {
	release, pre, _ := strings.Cut(v, "-")
	var nums []int
	for _, part := range strings.Split(release, ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			n = -1
		}
		nums = append(nums, n)
	}
	return nums, pre
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if goos == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}
