package manager

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

type Devices interface {
	Run(ctx context.Context, runSocketListener bool) error
	Devices() (map[string][]device.Device, error)
	SendAction(dev device.Device, a action.Interface) error
	// UnlockDevice(s Session) error
	// LockDevice(s Session, properties *device.Properties) (*DeviceLock, error)
	//GetDevice(id uint) (device.Device, error)
	GetDevice(id string) (device.Device, string)
	Start(getDevice device.Device) error
	Stop(getDevice device.Device) error
	AddHandler(handler device.Handler)
	StopObserver()
	FindDevice(params *app.Parameter) device.Device
	RegisterDevice(data device.RegisterData) (device.Device, error)
	Managers() []string
}
