package generic

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

const (
	AttributePin = "PIN"
)

var (
	DeviceLockedError   = fmt.Errorf("can't lock device, device is already locked")
	DeviceUnlockedError = fmt.Errorf("can't unlock device, device was not locked")
)

type Device struct {
	con           *device.Connection
	writer        device.LogWriter
	actionHandler []action.ActionHandler
	locked        bool
	Config        *models.Device
	ctx           context.Context
	cancel        context.CancelFunc
}

func (d *Device) GetConfig() *models.Device {
	return d.Config
}

func (d *Device) SetConfig(config *models.Device) {
	d.Config = config
}

func (d *Device) NewContext() context.Context {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	return d.ctx
}

func (d *Device) SetConnection(connection *device.Connection) {
	if connection != nil {
		d.Log("device", "Device Connected: DeviceID: %s SessionID: %v AppType: %v, Version: %s", connection.ConnectionParameter.DeviceID, connection.ConnectionParameter.SessionID, connection.ConnectionParameter.AppType, connection.ConnectionParameter.Version)
	} else {
		d.Log("device", "Device Disconnected")
	}
	d.con = connection
}

func (d *Device) Cancel() {
	if d.cancel != nil {
		c := d.cancel
		d.cancel = nil
		c()
	}
}

func (d *Device) Connection() *device.Connection {
	return d.con
}

func (d *Device) Send(data []byte) error {
	if err := d.con.Send(data); err != nil {
		return err
	}
	return nil
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
	d.Cancel()
	d.locked = false
	return nil
}

func (d *Device) IsLocked() bool {
	return d.locked
}

func (d *Device) SetLogWriter(writer device.LogWriter) {
	d.writer = writer
}

func (d *Device) LogPerformance(checkpoint string, cpu, fps, mem float64, other string) {
	if d.writer != nil {
		d.writer.LogPerformance(checkpoint, cpu, fps, mem, other)
	}
}

func (d *Device) Data(source, path string) {
	if d.writer != nil {
		d.writer.Data(source, path)
	}
}

func (d *Device) Log(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Log(source, format, params...)
	} else {
		logrus.Infof(format, params...)
	}
}

func (d *Device) Error(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Error(source, format, params...)
	} else {
		logrus.Errorf(format, params...)
	}
}

func (d *Device) Exception(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Error(source, format, params...)
	} else {
		logrus.Errorf(format, params...)
	}
}

func (d *Device) AddActionHandler(handler action.ActionHandler) {
	d.actionHandler = append(d.actionHandler, handler)
}

func (d *Device) RemoveActionHandler(handler action.ActionHandler) {
	for i := range d.actionHandler {
		if d.actionHandler[i] == handler {
			d.actionHandler = append(d.actionHandler[:i], d.actionHandler[i+1:]...)
		}
	}
}

func (d *Device) ActionHandlers() []action.ActionHandler {
	return d.actionHandler
}
