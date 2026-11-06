package device

type Handler interface {
	Name() string
	Init(masterUrl string, nodeIdentifier string, authToken *string) error
	Start() error
	Stop() error
	StartDevice(string) error
	StopDevice(string) error
	GetDevices() ([]Device, error)
	RefreshDevices(bool) error
	HasDevice(Device) bool
	RegisterDevice(data RegisterData) (Device, error)
}

// ChangeNotifying is implemented by handlers that detect device changes on
// their own (adb or usbmuxd events). The device manager hands them a callback
// to report a change, so a node pushes it to the master right away instead
// of with its next local refresh.
type ChangeNotifying interface {
	SetChangeNotifier(notify func())
}
