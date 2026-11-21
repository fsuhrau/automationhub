package iosdevice

import (
	"fmt"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

// The runner of the go-ios launcher (iOS before 17).
const (
	wdaBundleID     = "com.automationhub.WebDriverAgentRunner.xctrunner"
	wdaXCTestConfig = "WebDriverAgentRunner.xctest"
	aliveTimeout    = 3 * time.Second
)

// StartXCUITestRunner starts WebDriverAgent and creates a session. A runner
// that is already up (e.g. started for the live video) is reused.
func (d *Device) StartXCUITestRunner() error {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	return d.startWebDriver()
}

// StopXCUITestRunner closes the session and ends WebDriverAgent. The device
// stays attached; its state follows usbmuxd, not WebDriverAgent.
func (d *Device) StopXCUITestRunner() error {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	if d.webDriver != nil {
		_ = d.webDriver.CloseSession()
	}
	d.stopRunner()
	return nil
}

// ensureWebDriver starts WebDriverAgent unless it is up already. It returns
// the client and the URL of the MJPEG server.
func (d *Device) ensureWebDriver() (*webdriver.Client, string, error) {
	d.wdaMu.Lock()
	defer d.wdaMu.Unlock()
	if wd := d.webDriver; wd == nil || !wd.Alive(aliveTimeout) {
		if err := d.startWebDriver(); err != nil {
			return nil, "", err
		}
	}
	if d.wdaRunner != nil {
		return d.webDriver, d.wdaRunner.MjpegAddress, nil
	}
	mjpeg, err := d.webDriver.MjpegURL(webdriver.DefaultMjpegServerPort)
	return d.webDriver, mjpeg, err
}

// stopRunner ends WebDriverAgent; the caller holds wdaMu.
func (d *Device) stopRunner() {
	if d.wdaRunner != nil {
		d.wdaRunner.Stop()
		d.wdaRunner = nil
	} else {
		_ = d.wda.deps.USB.CloseXCUI()
	}
	d.webDriver = nil
}

// startWebDriver opens a session on the running WebDriverAgent, or starts one
// first; the caller holds wdaMu.
func (d *Device) startWebDriver() error {
	if wd := d.webDriver; wd != nil && wd.Alive(aliveTimeout) {
		return d.createSession(wd)
	}
	if d.webDriver != nil || d.wdaRunner != nil {
		// a runner that stopped answering
		d.stopRunner()
	}
	d.screenScale = 0
	address, err := d.launchRunner()
	if err != nil {
		return err
	}
	d.webDriver = webdriver.New(address, d.wda.deps.HTTP)
	return d.createSession(d.webDriver)
}

func (d *Device) createSession(wd *webdriver.Client) error {
	if _, err := wd.CreateSession(); err != nil {
		return fmt.Errorf("create WebDriverAgent session: %w", err)
	}
	return nil
}

// launchRunner starts WebDriverAgent with the launcher that fits the iOS
// version and returns its URL.
func (d *Device) launchRunner() (string, error) {
	if useXcodebuild(d.wda.cfg, d.OSVersion) {
		return d.launchXcodebuild()
	}
	return d.launchGoIOS()
}

// launchXcodebuild attaches to a running WebDriverAgent, or starts one with
// xcodebuild.
func (d *Device) launchXcodebuild() (string, error) {
	runner, attachErr := d.wda.attach(d.ID)
	if attachErr == nil {
		d.logf("Use the WebDriverAgent that already runs on the device")
	} else {
		d.logf("No running WebDriverAgent: %v", attachErr)
		var err error
		if runner, err = d.wda.startXcodebuild(d.ID); err != nil {
			return "", fmt.Errorf("no running WebDriverAgent (%v), and starting one failed: %w", attachErr, err)
		}
	}
	d.wdaRunner = runner
	return runner.Address, nil
}

// launchGoIOS runs the WebDriverAgentRunner through testmanagerd and takes its
// URL from the go-ios log output.
func (d *Device) launchGoIOS() (string, error) {
	entry, err := d.wda.deps.USB.GetDevice(d.ID)
	if err != nil {
		return "", err
	}
	return d.wda.deps.Hook.Await(WDAStartTimeout, d.wda.deps.Clock.After, func(fail func(error)) {
		go d.runXCUI(entry, fail)
	})
}

func (d *Device) runXCUI(entry ios.DeviceEntry, fail func(error)) {
	err := d.wda.deps.USB.RunXCUI(wdaBundleID, wdaBundleID, wdaXCTestConfig, entry, nil, nil)
	if err != nil {
		d.Error("device", "WebDriverAgent runner stopped: %v", err)
		fail(err)
	}
}
