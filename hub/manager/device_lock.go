package manager

import (
	"github.com/fsuhrau/automationhub/device"
)

type DeviceLock struct {
	Device       device.Device
	AppName      string
}