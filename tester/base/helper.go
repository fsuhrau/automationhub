package base

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"path/filepath"
)

func GetParams(binary *models.AppBinary, startupUrl string) app.Parameter {
	var params app.Parameter
	if binary != nil {
		// handle parameter with binary
		params.Identifier = binary.App.Identifier
		params.Name = binary.Name
		params.Version = binary.Version
		var android *app.AndroidParams
		if binary.Android.LaunchActivity != "" {
			android = &app.AndroidParams{
				LaunchActivity: binary.Android.LaunchActivity,
			}
		}
		var executable *app.ExecutableParams
		if binary.Executable.Executable != "" {
			executable = &app.ExecutableParams{
				Executable: binary.Executable.Executable,
			}
		}
		params.App = &app.AppParams{
			AppBinaryID: binary.ID,
			AppPath:     filepath.Join(apps.AppStoragePath, binary.AppPath),
			Hash:        binary.Hash,
			Size:        binary.Size,
			Additional:  binary.Additional,
			Android:     android,
			Executable:  executable,
		}
	} else if startupUrl != "" {
		params.Web = &app.WebParams{
			StartURL: startupUrl,
		}
	} else {
		params.Web = &app.WebParams{
			StartURL: startupUrl,
		}
	}
	return params
}
