package action

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	hubsync "github.com/fsuhrau/automationhub/utils/sync"
)

// ErrDeviceDisconnected is returned when the app disconnects before it
// answered the action.
var ErrDeviceDisconnected = errors.New("device disconnected before action completed")

type actionExecutor struct {
	devicesManager manager.Devices

	request      action.Interface
	done         chan struct{}
	once         sync.Once
	disconnected atomic.Bool
}

func NewExecutor(devices manager.Devices) *actionExecutor {
	return &actionExecutor{
		devicesManager: devices,
	}
}

// Call is one action for a device and how long the app has to answer it.
type Call struct {
	Device  device.Device
	Action  action.Interface
	Timeout time.Duration
}

// Execute sends the action to the device and waits for the matching response.
// It returns when the response arrived, the app disconnected, the context was
// cancelled or the timeout elapsed. A late response after Execute returned is
// ignored and never blocks the caller delivering it.
func (e *actionExecutor) Execute(ctx context.Context, call Call) error {
	// reset state; the executor is not registered yet, so no concurrent access
	e.request = call.Action
	e.done = make(chan struct{})
	e.once = sync.Once{}
	e.disconnected.Store(false)

	call.Device.AddActionHandler(e)
	defer call.Device.RemoveActionHandler(e)

	if err := e.devicesManager.SendAction(call.Device, call.Action); err != nil {
		return err
	}

	timer := time.NewTimer(call.Timeout)
	defer timer.Stop()

	select {
	case <-e.done:
		if e.disconnected.Load() {
			return ErrDeviceDisconnected
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return hubsync.ErrTimeout
	}
}

func (e *actionExecutor) finish(fn func()) {
	e.once.Do(func() {
		if fn != nil {
			fn()
		}
		close(e.done)
	})
}

func (e *actionExecutor) OnActionResponse(d interface{}, response *appproto.Response) {
	if response == nil {
		e.disconnected.Store(true)
		e.finish(nil)
		if dev, ok := d.(device.Device); ok {
			dev.Error("testrunner", "Device Disconnected")
		}
		return
	}

	if response.ActionType == e.request.GetActionType() {
		e.finish(func() {
			// an action that cannot read its response stays without content;
			// its caller sees that in the action
			_ = e.request.ProcessResponse(response)
		})
	}
}
