package tester

import (
	"github.com/fsuhrau/automationhub/storage/models"
)

type Interface interface {
	Initialize(test models.Test, env map[string]string) error
	Run(devs []models.Device, appData *models.App) (*models.TestRun, error)
}
