package iossim

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	_, err := d.simctl.run("get_app_container", d.ID, params.Identifier)
	return err == nil, err
}

func (d *Device) InstallApp(params *app.Parameter) error {
	_, err := d.simctl.run("install", d.ID, params.App.AppPath)
	return err
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	_, err := d.simctl.run("uninstall", d.ID, params.Identifier)
	return err
}

//nolint:revive // argument-limit: signature is fixed by device.AppManager
func (d *Device) StartApp(_ *device.DeviceConfig, params *app.Parameter, sessionID, nodeURL string) error {
	if d.deps.KeepAppRunning {
		return nil
	}
	_, err := d.simctl.run("launch", d.ID, params.Identifier, "SESSION_ID", sessionID, "NODE_URL", nodeURL)
	return err
}

func (d *Device) StopApp(params *app.Parameter) error {
	if d.deps.KeepAppRunning {
		return nil
	}
	_, err := d.simctl.run("terminate", d.ID, params.Identifier)
	return err
}
