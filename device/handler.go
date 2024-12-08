package device

type Handler interface {
	Name() string
	Init(masterUrl string, nodeIdentifier string) error
	Start() error
	Stop() error
	StartDevice(string) error
	StopDevice(string) error
	GetDevices() ([]Device, error)
	RefreshDevices(bool) error
	HasDevice(Device) bool
	RegisterDevice(data RegisterData) (Device, error)
}
