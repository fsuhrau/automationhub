package macos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsuhrau/automationhub/app"
)

func (d *Device) appRoot(params *app.Parameter) string {
	return filepath.Join(d.deps.InstallDir, params.App.Hash)
}

func (d *Device) executablePath(params *app.Parameter) string {
	return filepath.Join(d.appRoot(params), params.Name, params.App.Executable.Executable)
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	info, err := os.Stat(filepath.Join(d.appRoot(params), params.Name))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat app: %w", err)
	}
	return info.IsDir(), nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	if _, err := app.Unzip(params.App.AppPath, d.appRoot(params)); err != nil {
		return fmt.Errorf("unzip app: %w", err)
	}
	if err := os.Chmod(d.executablePath(params), 0o755); err != nil {
		return fmt.Errorf("make app executable: %w", err)
	}
	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	if err := os.RemoveAll(d.appRoot(params)); err != nil {
		return fmt.Errorf("remove app: %w", err)
	}
	return nil
}

// processArgs are the arguments the app is started with. The bare NODE_URL
// argument is what the app has always been given.
func processArgs(sessionID, nodeURL, deviceID string) []string {
	return []string{"--sessionId=" + sessionID, "NODE_URL", "--nodeURL=" + nodeURL, "--deviceId=" + deviceID}
}

func (d *Device) launch(executable string, args []string) error {
	// The app outlives the call, so it does not run on a request context.
	p, err := d.deps.Runner.Start(context.Background(), executable, args...)
	if err != nil {
		return fmt.Errorf("start app: %w", err)
	}
	go func() { _ = p.Wait() }() // reap the process when the app exits
	return nil
}
