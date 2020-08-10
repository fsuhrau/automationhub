package device

import "errors"

var (
	DeviceNotFoundError  = errors.New("Device not found")
	ManagerNotFoundError = errors.New("Manager not found")
)
