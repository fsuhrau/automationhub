package base

import (
	"path/filepath"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// GetParams builds the start parameters of an app: the bundle when there is
// one, else the web page to open.
func GetParams(storeDir string, binary *models.AppBinary, startupUrl string) app.Parameter {
	if binary == nil {
		return app.Parameter{Web: &app.WebParams{StartURL: startupUrl}}
	}
	return bundleParams(storeDir, binary)
}

// bundleParams describes the bundle to install and start.
func bundleParams(storeDir string, binary *models.AppBinary) app.Parameter {
	params := app.Parameter{Identifier: bundleIdentifier(binary), Name: binary.Name, Version: binary.Version}
	params.App = &app.AppParams{
		AppBinaryID: binary.ID,
		AppPath:     filepath.Join(storeDir, binary.AppPath),
		Hash:        binary.Hash,
		Size:        binary.Size,
		Additional:  binary.Additional,
		Android:     androidParams(binary),
		Executable:  executableParams(binary),
	}
	return params
}

// bundleIdentifier is the package of the bundle: feature and enterprise
// builds use another package than the app is registered with.
func bundleIdentifier(binary *models.AppBinary) string {
	if binary.Identifier == "" && binary.App != nil {
		return binary.App.Identifier
	}
	return binary.Identifier
}

func androidParams(binary *models.AppBinary) *app.AndroidParams {
	if binary.Android.LaunchActivity == "" {
		return nil
	}
	return &app.AndroidParams{LaunchActivity: binary.Android.LaunchActivity}
}

func executableParams(binary *models.AppBinary) *app.ExecutableParams {
	if binary.Executable.Executable == "" {
		return nil
	}
	return &app.ExecutableParams{Executable: binary.Executable.Executable}
}
