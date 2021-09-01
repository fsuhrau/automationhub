package tester

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

type Interface interface {
	Initialize(test models.Test) error
	Run(dev device.Device, app models.App)
}