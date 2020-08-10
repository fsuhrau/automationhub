package device

type Manager interface {
	Name() string
	Start() error
	Stop() error
	StartDevice(string) error
	StopDevice(string) error
	GetDevices() ([]Device, error)
	RefreshDevices() error

	HasDevice(Device) bool
}
