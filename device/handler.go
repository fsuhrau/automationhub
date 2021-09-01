package device

type DeviceUpdateFunc func(dev Device)

type Handler interface {
	Name() string
	Init() error
	Start() error
	Stop() error
	StartDevice(string) error
	StopDevice(string) error
	GetDevices() ([]Device, error)
	RefreshDevices(DeviceUpdateFunc) error

	HasDevice(Device) bool
}
