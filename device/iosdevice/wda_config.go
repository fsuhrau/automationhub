package iosdevice

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/config"
)

// Launchers of WebDriverAgent (config webdriver.launcher).
const (
	// LauncherAuto uses xcodebuild for iOS 17 and later, go-ios before.
	LauncherAuto = "auto"
	// LauncherXcodebuild runs the WebDriverAgentRunner with xcodebuild, which
	// mounts the developer disk image and sets up the iOS 17 tunnel itself.
	LauncherXcodebuild = "xcodebuild"
	// LauncherGoIOS runs it through testmanagerd with go-ios; up to iOS 16.
	LauncherGoIOS = "go-ios"

	defaultWDAScheme = "WebDriverAgentRunner"
	// XcodebuildStartTimeout bounds an xcodebuild start; "xcodebuild test"
	// may build first.
	XcodebuildStartTimeout = 120 * time.Second
	wdaDevicePort          = 8100
)

// useXcodebuild decides the launcher for an iOS version like "18.7.1".
func useXcodebuild(cfg *config.WebDriver, osVersion string) bool {
	launcher := LauncherAuto
	if cfg != nil && cfg.Launcher != "" {
		launcher = strings.ToLower(cfg.Launcher)
	}
	switch launcher {
	case LauncherXcodebuild:
		return true
	case LauncherGoIOS:
		return false
	}
	major, _ := strconv.Atoi(strings.SplitN(osVersion, ".", 2)[0])
	return major >= 17
}

// xcodebuildArgs runs a prebuilt runner (.xctestrun from "xcodebuild
// build-for-testing") or builds and runs the WebDriverAgent project.
func xcodebuildArgs(cfg *config.WebDriver, udid string) ([]string, error) {
	if cfg == nil || (cfg.XCTestRun == "" && cfg.Project == "") {
		return nil, errors.New("iOS 17 and later start WebDriverAgent with xcodebuild: set webdriver.xctestrun (a WebDriverAgentRunner .xctestrun) or webdriver.project (WebDriverAgent.xcodeproj) of the ios_device manager on this host")
	}
	destination := "id=" + udid
	var args []string
	if cfg.XCTestRun != "" {
		args = []string{"test-without-building", "-xctestrun", cfg.XCTestRun, "-destination", destination}
	} else {
		scheme := cfg.Scheme
		if scheme == "" {
			scheme = defaultWDAScheme
		}
		args = []string{"test", "-project", cfg.Project, "-scheme", scheme, "-destination", destination, "-allowProvisioningUpdates"}
	}
	if cfg.DerivedData != "" {
		args = append(args, "-derivedDataPath", cfg.DerivedData)
	}
	return args, nil
}
