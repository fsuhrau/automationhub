package web

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// A web app is a page, there is nothing to install: it is always installed
// and installing or uninstalling it succeeds without doing anything.

func (d *Device) IsAppInstalled(*app.Parameter) (bool, error) { return true, nil }

func (d *Device) InstallApp(*app.Parameter) error { return nil }

func (d *Device) UninstallApp(*app.Parameter) error { return nil }

// StartApp opens the start url, extended by the session, in the browser.
//
//nolint:revive // argument-limit: signature is fixed by device.AppManager
func (d *Device) StartApp(_ *device.DeviceConfig, params *app.Parameter, sessionID, nodeURL string) error {
	url := d.applicationURL(params, sessionID, nodeURL)
	// The browser outlives the call, so it does not run on a request context.
	p, err := d.env.Runner.Start(context.Background(), d.browserPath, url)
	if err != nil {
		return fmt.Errorf("start browser: %w", err)
	}
	d.process = p
	return nil
}

// StopApp kills the browser that StartApp started; without one it does nothing.
func (d *Device) StopApp(*app.Parameter) error {
	p := d.process
	d.process = nil
	if p == nil {
		return nil
	}
	if err := p.Kill(); err != nil {
		return fmt.Errorf("stop browser: %w", err)
	}
	return nil
}
