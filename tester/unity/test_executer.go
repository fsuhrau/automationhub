package unity

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	"time"
)

type testExecuter struct {
	devicesManager manager.Devices
	fin            chan bool
}

func NewExecuter(devices manager.Devices) *testExecuter {
	return &testExecuter{
		devicesManager: devices,
		fin: make(chan bool, 1),
	}
}

func (e *testExecuter) Execute(dev device.Device, test action.TestStart, timeout time.Duration) error {
	dev.SetActionHandler(e)

	finishWaitingGroup := sync.ExtendedWaitGroup{}
	finishWaitingGroup.Add(1)
	go func(wg *sync.ExtendedWaitGroup) {
		select {
		case finished := <-e.fin:
			{
				if finished {
					logrus.Debug("test finished")
					wg.Done()
					break
				}
			}
		}
	}(&finishWaitingGroup)
	if err := e.devicesManager.SendAction(dev, &test); err != nil {
		dev.Error(err.Error())
		return err
	}

	if err := finishWaitingGroup.WaitWithTimeout(timeout); err != nil {
		dev.Error(err.Error())
		return err
	}
	return nil
}

func (tr *testExecuter) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		tr.fin <- true
		dev.Log("testsRunner Device Disconnected")
		return
	}

	dev.Log("testsRunner Action Response: %v", response)
	if response.ActionType == action.ActionType_Log {
		if response.GetValue() == "End" {
			tr.fin <- true
		}
	}
}
