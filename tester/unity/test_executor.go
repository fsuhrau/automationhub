package unity

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/utils/sync"
	"time"
)

type testExecutor struct {
	devicesManager manager.Devices
	fin            chan bool
	wg sync.ExtendedWaitGroup
}

func NewExecutor(devices manager.Devices) *testExecutor {
	return &testExecutor{
		devicesManager: devices,
		fin:            make(chan bool, 1),
	}
}

func (e *testExecutor) Execute(dev device.Device, test action.TestStart, timeout time.Duration) error {
	dev.AddActionHandler(e)
	defer func () {
		dev.RemoveActionHandler(e)
	}()

	e.wg = sync.ExtendedWaitGroup{}
	e.wg.Add(1)
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
	}(&e.wg)
	if err := e.devicesManager.SendAction(dev, &test); err != nil {
		dev.Error("testrunner", err.Error())
		return err
	}

	if err := e.wg.WaitUntil(time.Now().Add(30 * time.Second)); err != nil {
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
		if response.Success {
			dev.Log("testrunner", fmt.Sprintf("timeout: %d", response.GetTestDetails().Timeout))
			dev.Log("testrunner", fmt.Sprintf("categories: %v", response.GetTestDetails().Categories))
			if response.GetTestDetails().Timeout > 0 {
				tr.wg.UpdateUntil(time.Now().Add(time.Duration(response.GetTestDetails().Timeout) * time.Millisecond))
			}
		} else {
			dev.Error("testrunner", "starting test failed")
			tr.fin <- true
			return
		}
	}

	if response.ActionType == action.ActionType_ExecutionResult {
		if response.Success {
			dev.Log("testrunner", "test finished successfully")

		} else {
			dev.Error("testrunner", "test finished with errors")
		}
		tr.fin <- true
		return
	}

	// logrus.Info(response)
}
