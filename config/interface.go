package config

type Interface interface {
	GetDeviceConfig(string) *Device
	GetDevicesForManager(string) []Device
}
