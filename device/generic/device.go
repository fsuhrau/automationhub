package generic

import (
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
}

func (d *Device) GetConfig() *models.Device {
	return d.Config
}

func (d *Device) SetConfig(config *models.Device) {
	d.Config = config
}

func (d *Device) SetConnection(connection *device.Connection) {
	if connection != nil {
		d.Log("device", "Device Connected: DeviceID: %s SessionID: %v AppType: %v, Version: %s", connection.ConnectionParameter.DeviceID, connection.ConnectionParameter.SessionID, connection.ConnectionParameter.AppType, connection.ConnectionParameter.Version)
	} else {
		d.Log("device", "Device Disconnected")
	}
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

func (d *Device) IsLocked() bool {
	return d.locked
}

func (d *Device) SetLogWriter(writer device.LogWriter) {
	d.writer = writer
}

func (d *Device) LogPerformance(checkpoint string, cpu, fps, mem float32, other string) {
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

/*
	data := <-dev.Connection().ResponseChannel
	resp := &action.Response{}
	if err := proto.Unmarshal(data.Data, resp); err != nil {
		dm.log.Errorf("SocketAccept ReadError: %v", err)
	}
	logrus.Infof("send bytes response: %v", resp)

*/
/*
func (d *Device) SendAction(act action.Interface) error {
	start := time.Now()
	defer func(t time.Time) {
		elapsed := time.Since(start)
		d.Log("Send Action took: %s", elapsed.String())
	}(start)

	d.Log("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)
	buf, err := act.Serialize()
	if err != nil {
		return fmt.Errorf("Could not marshal Action: %v", err)
	}
	if err := d.Connection().Send(buf); err != nil {
		return err
	}

	if err != nil {
		return err
	}
	dm.log.Debugf("ProcessResponse Action")
	if err := act.ProcessResponse(response); err != nil {
		return err
	}
	dev.Log("action response: %v", act)
	return nil
}
*/
