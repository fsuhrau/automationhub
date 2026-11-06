package generic

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	AttributePin = "PIN"
)

var (
	DeviceLockedError   = fmt.Errorf("can't lock device, device is already locked")
	DeviceUnlockedError = fmt.Errorf("can't unlock device, device was not locked")
	ErrNotConnected     = fmt.Errorf("app is not connected to the device")
)

type Device struct {
	mu            sync.Mutex
	con           *device.Connection
	writer        device.LogWriter
	actionHandler []action.ActionHandler
	locked        bool
	Config        *models.Device `json:"config"`
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
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ctx, d.cancel = context.WithCancel(context.Background())
	return d.ctx
}

func (d *Device) SetConnection(connection *device.Connection) {
	if connection != nil {
		d.Log("device", "Device Connected: DeviceID: %s SessionID: %v AppType: %v, Version: %s", connection.ConnectionParameter.DeviceID, connection.ConnectionParameter.SessionID, connection.ConnectionParameter.AppType, connection.ConnectionParameter.Version)
	} else {
		d.Log("device", "Device Disconnected")
	}
	d.mu.Lock()
	d.con = connection
	d.mu.Unlock()
}

func (d *Device) Cancel() {
	d.mu.Lock()
	c := d.cancel
	d.cancel = nil
	d.mu.Unlock()
	if c != nil {
		c()
	}
}

func (d *Device) Connection() *device.Connection {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.con
}

func (d *Device) Send(data []byte) error {
	con := d.Connection()
	if con == nil {
		return ErrNotConnected
	}
	return con.Send(data)
}

// Lock atomically marks the device as locked. It fails if it is locked already.
func (d *Device) Lock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.locked {
		return DeviceLockedError
	}
	d.locked = true
	return nil
}

func (d *Device) Unlock() error {
	d.mu.Lock()
	if !d.locked {
		d.mu.Unlock()
		return DeviceUnlockedError
	}
	d.locked = false
	c := d.cancel
	d.cancel = nil
	d.mu.Unlock()
	if c != nil {
		c()
	}
	return nil
}

// Release clears the lock without cancelling the device context, so a
// connected app stays connected. Used for interactive control leases.
func (d *Device) Release() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.locked {
		return DeviceUnlockedError
	}
	d.locked = false
	return nil
}

func (d *Device) IsLocked() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.locked
}

func (d *Device) SetLogWriter(writer device.LogWriter) {
	d.writer = writer
}

func (d *Device) GetLogWriter() device.LogWriter {
	return d.writer
}

func (d *Device) SetTestResult(success models.TestResultState) {

}

func (d *Device) TestProtocolId() *uint {
	if d.writer != nil {
		return d.writer.TestProtocolId()
	}
	return nil
}

func (d *Device) LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string) {
	if d.writer != nil {
		d.writer.LogPerformance(checkpoint, cpu, fps, mem, vertexCount, triangles, other)
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

func (d *Device) Passed(result bool) {
	if d.writer != nil {
		d.writer.Passed(result)
	} else {
		logrus.Infof("Test: %v", result)
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
	d.mu.Lock()
	defer d.mu.Unlock()
	d.actionHandler = append(d.actionHandler, handler)
}

func (d *Device) RemoveActionHandler(handler action.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.actionHandler {
		if d.actionHandler[i] == handler {
			handlers := make([]action.ActionHandler, 0, len(d.actionHandler)-1)
			handlers = append(handlers, d.actionHandler[:i]...)
			d.actionHandler = append(handlers, d.actionHandler[i+1:]...)
			return
		}
	}
}

// ActionHandlers returns a snapshot of the registered handlers that is safe to
// iterate while handlers are added or removed.
func (d *Device) ActionHandlers() []action.ActionHandler {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]action.ActionHandler(nil), d.actionHandler...)
}
