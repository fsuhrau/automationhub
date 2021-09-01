package generic

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
)

var (
	DeviceLockedError = fmt.Errorf("can't lock device, device is already locked")
	DeviceUnlockedError = fmt.Errorf("can't unlock device, device was not locked")
)

type Device struct {
	con *device.Connection
	locked bool
}

func (d *Device) SetConnection(connection *device.Connection) {
	d.con = connection
}

func (d *Device) Connection() *device.Connection {
	return d.con
}

func (d *Device) Lock() error {
	if d.locked {
		return DeviceLockedError
	}
	d.locked = true
	return nil
}

func (d *Device) Unlock() error {
	if !d.locked {
		return DeviceUnlockedError
	}
	d.locked = false
	return nil
}

func (d *Device) IsLocker() bool {
	return d.locked
}
