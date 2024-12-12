package action

import (
	"context"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"time"
)

type actionExecutor struct {
	devicesManager manager.Devices
	fin            chan bool
	request        action.Interface
}

func NewExecutor(devices manager.Devices) *actionExecutor {
	return &actionExecutor{
		devicesManager: devices,
		fin:            make(chan bool, 1),
	}
}

func (e *actionExecutor) Execute(ctx context.Context, dev device.Device, a action.Interface, timeout time.Duration) error {
	e.request = a
	dev.AddActionHandler(e)
	defer func() {
		dev.RemoveActionHandler(e)
	}()

	finishWaitingGroup := sync.NewExtendedWaitGroup(ctx)
	finishWaitingGroup.Add(1)
	go func(wg sync.ExtendedWaitGroup) {
		select {
		case finished := <-e.fin:
			{
				if finished {
					wg.Done()
					break
				}
			}
		}
	}(finishWaitingGroup)
	if err := e.devicesManager.SendAction(dev, a); err != nil {
		return err
	}

	if err := finishWaitingGroup.WaitWithTimeout(timeout); err != nil {
		return err
	}

	return nil
}

func (e *actionExecutor) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		e.fin <- true
		dev.Error("testrunner", "Device Disconnected")
		return
	}

	if response.ActionType == e.request.GetActionType() {
		e.request.ProcessResponse(response)
		e.fin <- true
	}
}
