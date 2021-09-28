package unity

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"time"
)

type testExecutor struct {
	devicesManager manager.Devices
	fin            chan bool
}

func NewExecutor(devices manager.Devices) *testExecutor {
	return &testExecutor{
		devicesManager: devices,
		fin:            make(chan bool, 1),
	}
}

func (e *testExecutor) Execute(dev device.Device, test action.TestStart, timeout time.Duration) error {
	dev.SetActionHandler(e)

	finishWaitingGroup := sync.ExtendedWaitGroup{}
	finishWaitingGroup.Add(1)
	go func(wg *sync.ExtendedWaitGroup) {
		select {
		case finished := <-e.fin:
			{
				if finished {
					wg.Done()
					break
				}
			}
		}
	}(&finishWaitingGroup)
	if err := e.devicesManager.SendAction(dev, &test); err != nil {
		dev.Error("testrunner", err.Error())
		return err
	}

	if err := finishWaitingGroup.WaitWithTimeout(timeout); err != nil {
		dev.Error("testrunner", err.Error())
		return err
	}
	return nil
}

func (tr *testExecutor) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		tr.fin <- true
		dev.Error("testrunner", "Device Disconnected")
		return
	}

	if response.ActionType == action.ActionType_ExecuteTest {
		if !response.Success {
			dev.Error("testrunner", "starting test failed")
			tr.fin <- true
		}
	}

	if response.ActionType == action.ActionType_ExecutionResult {
		if response.Success {
			dev.Log("testrunner", "test finished successfully")

		} else {
			dev.Error("testrunner", "test finished with errors")
		}
		tr.fin <- true

	}
		/*
		if response.ActionType == action.ActionType_Log {
			if response.GetLog().GetType() == action.LogType_StatusLog {
				if response.GetLog().Message == "End" {
					dev.Log("testrunner_status", "test finished")
					tr.fin <- true
				}
			}
		}
		 */
}
