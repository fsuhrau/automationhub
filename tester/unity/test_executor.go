package unity

import (
	"context"
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
	wg             sync.ExtendedWaitGroup
	actionHandler  map[action.ActionType]func(device.Device, *action.Response)
}

func NewExecutor(devices manager.Devices) *testExecutor {
	executor := &testExecutor{
		devicesManager: devices,
		fin:            make(chan bool, 1),
	}
	executor.actionHandler = map[action.ActionType]func(device.Device, *action.Response){
		action.ActionType_ExecuteTest:     executor.handleExecuteTest,
		action.ActionType_ExecutionResult: executor.handleExecutionResult,
		action.ActionType_Performance:     executor.handlePerformance,
		action.ActionType_NativeScript:    executor.handleNativeScript,
		action.ActionType_Log:             executor.handleLog,
	}
	return executor
}

func (e *testExecutor) Execute(ctx context.Context, dev device.Device, test action.TestStart, timeout time.Duration) error {
	dev.AddActionHandler(e)
	defer func() {
		dev.RemoveActionHandler(e)
	}()

	e.wg = sync.NewExtendedWaitGroup(ctx)
	e.wg.Add(1)
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
	}(e.wg)
	if err := e.devicesManager.SendAction(dev, &test); err != nil {
		dev.Error("testrunner", err.Error())
		return err
	}

	if err := e.wg.WaitUntil(time.Now().Add(timeout)); err != nil {
		dev.Error("testrunner", err.Error())
		return err
	}
	return nil
}

func (tr *testExecutor) handleExecuteTest(dev device.Device, response *action.Response) {
	if response.Success {
		dev.Log("testrunner", fmt.Sprintf("timeout: %d", response.Payload.TestDetails.Timeout))
		dev.Log("testrunner", fmt.Sprintf("categories: %v", response.Payload.TestDetails.Categories))
		if response.Payload.TestDetails.Timeout > 0 {
			tr.wg.UpdateUntil(time.Now().Add(time.Duration(response.Payload.TestDetails.Timeout) * time.Millisecond))
		}
	} else {
		dev.Error("testrunner", "starting test failed")
		tr.fin <- true
	}
}

func (tr *testExecutor) handleExecutionResult(dev device.Device, response *action.Response) {
	if response.Success {
		dev.Log("testrunner", "test finished successfully")
	} else {
		dev.Error("testrunner", "test finished with errors")
	}
	tr.fin <- true
}

func (tr *testExecutor) handlePerformance(dev device.Device, response *action.Response) {
	dev.LogPerformance(response.Payload.PerformanceData.Checkpoint, response.Payload.PerformanceData.CPU, response.Payload.PerformanceData.FPS, response.Payload.PerformanceData.Memory, "")
}

func (tr *testExecutor) handleNativeScript(dev device.Device, response *action.Response) {
	go dev.RunNativeScript(*response.Payload.Data)
}

func (tr *testExecutor) handleLog(dev device.Device, response *action.Response) {
	if response.Payload.LogData.Level == action.LogLevel_Exception {
		tr.fin <- true
	}
}

func (tr *testExecutor) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		tr.fin <- true
		dev.Error("testrunner", "Device Disconnected")
		return
	}

	if handler, ok := tr.actionHandler[response.ActionType]; ok {
		handler(dev, response)
	}
}
