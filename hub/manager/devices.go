package manager

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

// DeviceRunner controls the lifetime of the device observer.
type DeviceRunner interface {
	Run(ctx context.Context, runSocketListener bool) error
	StopObserver()
}

// DeviceLookup finds devices and lists the device handlers.
type DeviceLookup interface {
	Devices() (map[string][]device.Device, error)
	GetDevice(id string) (device.Device, string)
	FindDevice(params *app.Parameter) device.Device
	Managers() []string
}

// DevicePower starts and stops devices.
type DevicePower interface {
	Start(getDevice device.Device) error
	Stop(getDevice device.Device) error
}

// DeviceRegistry adds handlers and registers devices at them.
type DeviceRegistry interface {
	AddHandler(handler device.Handler)
	RegisterDevice(data device.RegisterData) (device.Device, error)
}

// ActionSender sends an action to the app on a device.
type ActionSender interface {
	SendAction(dev device.Device, a action.Interface) error
}

// Devices is the device manager, composed of its roles.
type Devices interface {
	DeviceRunner
	DeviceLookup
	DevicePower
	DeviceRegistry
	ActionSender
}
