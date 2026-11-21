package androiddevice

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// IsAppInstalled implements device.AppManager: the app is installed and the
// installed build is the one this node installed last.
func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	hash, known := d.installedApps[params.Identifier]
	sameBuild := known && hash == params.App.Hash
	out, err := d.run("shell", "pm", "list", "packages")
	installed := strings.Contains(string(out), params.Identifier)
	d.Log("device", "App '%s' is installed: %t", params.Identifier, installed)
	return installed && sameBuild, err
}

// InstallApp implements device.AppManager.
func (d *Device) InstallApp(params *app.Parameter) error {
	d.Log("device", "Install App '%s'", params.Identifier)
	if _, err := d.run(d.installArgs(params.App.AppPath)...); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}
	d.installedApps[params.Identifier] = params.App.Hash
	return nil
}

// installArgs are the adb arguments that install the apk. Devices before
// Android 7 (API 24) take the combined flag, with "d" for debuggable apks.
func (d *Device) installArgs(apkPath string) []string {
	if d.apiLevel < 24 {
		debug := ""
		if d.isDebuggable(apkPath) {
			debug = "d"
		}
		return []string{"install", "-t", "-rg" + debug, apkPath}
	}
	return []string{"install", "-t", "-r", "-g", apkPath}
}

// isDebuggable asks aapt whether the apk is debuggable.
func (d *Device) isDebuggable(apkPath string) bool {
	aapt, err := d.deps.AAPT()
	if err != nil {
		return false
	}
	out, err := d.deps.Runner.Run(context.Background(), aapt, "dump", "badging", apkPath)
	return err == nil && strings.Contains(string(out), "application-debuggable")
}

// UninstallApp implements device.AppManager.
func (d *Device) UninstallApp(params *app.Parameter) error {
	d.Log("device", "Uninstall App '%s'", params.Identifier)
	if _, err := d.run("uninstall", params.Identifier); err != nil {
		return err
	}
	d.installedApps[params.Identifier] = ""
	return nil
}

// StartApp implements device.AppManager.
//
//nolint:revive // argument-limit: the signature is fixed by device.AppManager
func (d *Device) StartApp(config *device.DeviceConfig, params *app.Parameter, sessionID string, nodeURL string) error {
	d.Log("device", "Start App '%s' with Session: '%s'", params.Identifier, sessionID)
	if err := d.unlockScreen(config); err != nil {
		return err
	}
	if err := d.reverseNodePort(nodeURL); err != nil {
		return err
	}
	activity := fmt.Sprintf("%s/%s", params.Identifier, params.App.Android.LaunchActivity)
	_, err := d.run("shell", "am", "start", "-n", activity,
		"-e", "SESSION_ID", sessionID, "-e", "NODE_URL", nodeURL, "-e", "DEVICE_ID", d.ID)
	return err
}

// StopApp implements device.AppManager.
func (d *Device) StopApp(params *app.Parameter) error {
	d.Log("device", "Stop App '%s'", params.Identifier)
	_, err := d.run("shell", "am", "force-stop", params.Identifier)
	return err
}

// ClearAppData stops the app and wipes its data (adb shell pm clear).
func (d *Device) ClearAppData(params *app.Parameter) error {
	d.Log("device", "Clear App Data '%s'", params.Identifier)
	_, err := d.run("shell", "pm", "clear", params.Identifier)
	return err
}

// reverseNodePort makes a loopback node URL work on the phone. On the device,
// localhost is the device itself, so "adb reverse" forwards the node port from
// the device back to this machine. Other hosts are left alone.
func (d *Device) reverseNodePort(nodeURL string) error {
	port, ok := loopbackPort(nodeURL)
	if !ok {
		return nil
	}
	d.Log("device", "Reverse tcp:%s to the node", port)
	spec := "tcp:" + port
	if _, err := d.run("reverse", spec, spec); err != nil {
		return fmt.Errorf("adb reverse %s for node url %s: %w", spec, nodeURL, err)
	}
	return nil
}

// loopbackPort returns the port of nodeURL when its host is a loopback address.
func loopbackPort(nodeURL string) (string, bool) {
	u, err := url.Parse(nodeURL)
	if err != nil || !isLoopback(u.Hostname()) {
		return "", false
	}
	if port := u.Port(); port != "" {
		return port, true
	}
	return defaultPort(u.Scheme)
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func defaultPort(scheme string) (string, bool) {
	switch scheme {
	case "ws", "http":
		return "80", true
	case "wss", "https":
		return "443", true
	}
	return "", false
}
