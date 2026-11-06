package action

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	hubsync "github.com/fsuhrau/automationhub/utils/sync"
)

// ErrDeviceDisconnected is returned when the app disconnects before it
// answered the action.
var ErrDeviceDisconnected = errors.New("device disconnected before action completed")

// TimeoutError is returned when the action was not answered in time.
var TimeoutError = hubsync.TimeoutError

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

// Execute sends the action to the device and waits for the matching response.
// It returns when the response arrived, the app disconnected, the context was
// cancelled or the timeout elapsed. A late response after Execute returned is
// ignored and never blocks the caller delivering it.
func (e *actionExecutor) Execute(ctx context.Context, dev device.Device, a action.Interface, timeout time.Duration) error {
	// reset state; the executor is not registered yet, so no concurrent access
	e.request = a
	e.done = make(chan struct{})
	e.once = sync.Once{}
	e.disconnected.Store(false)

	dev.AddActionHandler(e)
	defer dev.RemoveActionHandler(e)

	if err := e.devicesManager.SendAction(dev, a); err != nil {
		return err
	}

	timer := time.NewTimer(timeout)
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
		return TimeoutError
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

func (e *actionExecutor) OnActionResponse(d interface{}, response *action.Response) {
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
			e.request.ProcessResponse(response)
		})
	}
}
