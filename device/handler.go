package device

type Handler interface {
	Name() string
	Init() error
	Start() error
	Stop() error
	StartDevice(string) error
	StopDevice(string) error
	GetDevices() ([]Device, error)
	RefreshDevices() error

	HasDevice(Device) bool
}
