package macos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

//nolint:revive // argument-limit: signature is fixed by device.AppManager
func (d *Device) StartApp(_ *device.DeviceConfig, params *app.Parameter, sessionID, nodeURL string) error {
	executable := d.executablePath(params)
	d.runningExecutable = filepath.Base(executable)
	return d.launch(executable, processArgs(sessionID, nodeURL, d.DeviceID()))
}

func (d *Device) StopApp(params *app.Parameter) error {
	d.runningExecutable = ""
	name := filepath.Base(params.App.Executable.Executable)
	if _, err := d.deps.Runner.Run(context.Background(), "killall", name); err != nil {
		return fmt.Errorf("stop app: %w", err)
	}
	return nil
}

// GetScreenshot captures the window of the running app with screencapture.
func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	ctx := context.Background()
	id, err := d.windowID(ctx)
	if err != nil {
		return nil, 0, 0, err
	}
	file := filepath.Join(d.deps.ScreenshotDir, id+".png")
	if _, err := d.deps.Runner.Run(ctx, "screencapture", "-l", id, file); err != nil {
		return nil, 0, 0, fmt.Errorf("screencapture: %w", err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read screenshot: %w", err)
	}
	return data, 0, 0, nil
}

// windowID asks the app for the id of its first window.
func (d *Device) windowID(ctx context.Context) (string, error) {
	script := fmt.Sprintf("tell application \"%s\" to id of window 1", d.runningExecutable)
	out, err := d.deps.Runner.Run(ctx, "osascript", "-e", script)
	if err != nil {
		return "", fmt.Errorf("window id: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
