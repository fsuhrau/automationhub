package macos

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsuhrau/automationhub/device/hostinfo"
)

// Deps are the injected collaborators and directories of the macOS platform.
// Unset fields get the production defaults.
type Deps struct {
	hostinfo.Env
	// InstallDir holds the unpacked apps, default ~/.automationhub/installs.
	InstallDir string
	// ScreenshotDir holds the screencapture files, default <tmp>/automation_hub.
	ScreenshotDir string
}

// withDefaults fills in every unset field.
func (d Deps) withDefaults() Deps {
	d.Env = d.WithDefaults()
	if d.InstallDir == "" {
		d.InstallDir = filepath.Join(os.Getenv("HOME"), ".automationhub", "installs")
	}
	if d.ScreenshotDir == "" {
		d.ScreenshotDir = filepath.Join(os.TempDir(), "automation_hub")
	}
	return d
}

// ensureDirs creates the install and screenshot directories.
func (d Deps) ensureDirs() error {
	for _, dir := range []string{d.InstallDir, d.ScreenshotDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
