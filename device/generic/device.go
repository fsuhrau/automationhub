package generic

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/storage/models"
	"sync"
)

const (
	AttributePin = "PIN"
)

//nolint:gochecknoglobals,staticcheck // sentinel errors are part of the package API; ErrFoo names are used by the platforms
var (
	DeviceLockedError   = fmt.Errorf("can't lock device, device is already locked")
	DeviceUnlockedError = fmt.Errorf("can't unlock device, device was not locked")
	ErrNotConnected     = fmt.Errorf("app is not connected to the device")
)

type Device struct {
	mu            sync.Mutex
	con           *device.Connection
	writer        device.LogWriter
	actionHandler []appproto.ActionHandler
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

func (d *Device) AddActionHandler(handler appproto.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.actionHandler = append(d.actionHandler, handler)
}

func (d *Device) RemoveActionHandler(handler appproto.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.actionHandler {
		if d.actionHandler[i] == handler {
			handlers := make([]appproto.ActionHandler, 0, len(d.actionHandler)-1)
			handlers = append(handlers, d.actionHandler[:i]...)
			d.actionHandler = append(handlers, d.actionHandler[i+1:]...)
			return
		}
	}
}

// ActionHandlers returns a snapshot of the registered handlers that is safe to
// iterate while handlers are added or removed.
func (d *Device) ActionHandlers() []appproto.ActionHandler {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]appproto.ActionHandler(nil), d.actionHandler...)
}
