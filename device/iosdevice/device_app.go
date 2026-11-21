package iosdevice

import (
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func (d *Device) logf(format string, args ...any) {
	d.Log("device", format, args...)
}

// StartApp launches the app through WebDriverAgent.
//
//nolint:revive // argument-limit: the signature is fixed by device.AppManager
func (d *Device) StartApp(_ *device.DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error {
	if d.webDriver == nil {
		return fmt.Errorf("webdriver not connected")
	}
	return d.webDriver.Launch(appParams.Identifier, true, []string{"SESSION_ID", sessionId, "DEVICE_ID", d.ID, "NODE_URL", nodeUrl})
}

// StopApp terminates the app when it is in the foreground.
func (d *Device) StopApp(params *app.Parameter) error {
	if d.webDriver == nil {
		return fmt.Errorf("webdriver not connected")
	}
	info, err := d.webDriver.ActiveAppInfo()
	if err != nil {
		return err
	}
	if info.Value.BundleId == params.Identifier {
		return d.webDriver.Terminate(params.Identifier)
	}
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

// RunNativeScript runs a hub script through WebDriverAgent.
func (d *Device) RunNativeScript(script []byte) {
	if d.webDriver == nil {
		d.Error("device", "Script Failed: no WebDriverAgent session for device %s", d.ID)
		return
	}
	scriptHandler := New(d.webDriver, d)
	if err := scriptHandler.Execute(string(script)); err != nil {
		d.Error("device", "Script Failed: %v", err)
		_ = d.webDriver.PressButton(KEYCODE_HOME)
		_ = d.webDriver.CloseSession()
	}
}
