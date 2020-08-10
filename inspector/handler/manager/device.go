package manager

import "github.com/fsuhrau/automationhub/device"

type DeviceManager interface {
	Devices() ([]device.Device, error)
}
