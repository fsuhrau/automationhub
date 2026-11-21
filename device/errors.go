package device

import "errors"

//nolint:gochecknoglobals,staticcheck // sentinel errors are part of the package API; ErrFoo names are used by the platforms
var (
	DeviceNotFoundError  = errors.New("Device not found")
	ManagerNotFoundError = errors.New("Manager not found")
)
