package unity

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"github.com/fsuhrau/automationhub/utils/sync"
	gosync "sync"
	"time"
)

type testExecutor struct {
	devicesManager manager.Devices
	finishOnce     gosync.Once
	wg             sync.ExtendedWaitGroup
	actionHandler  map[action.ActionType]func(device.Device, *action.Response)
	protocolWriter *protocol.ProtocolWriter
}

func NewExecutor(devices manager.Devices, protocolWriter *protocol.ProtocolWriter) *testExecutor {
	executor := &testExecutor{
		devicesManager: devices,
		protocolWriter: protocolWriter,
	}
	executor.actionHandler = map[action.ActionType]func(device.Device, *action.Response){
		action.ActionType_ExecuteTest:           executor.handleExecuteTest,
		action.ActionType_ExecutionResult:       executor.handleExecutionResult,
		action.ActionType_Performance:           executor.handlePerformance,
		action.ActionType_NativeScript:          executor.handleNativeScript,
		action.ActionType_Log:                   executor.handleLog,
		action.ActionType_ExecuteMethodStart:    executor.handleTestMethodStart,
		action.ActionType_ExecuteMethodFinished: executor.handleTestMethodFinished,
	}
	return executor
}

func (e *testExecutor) Execute(ctx context.Context, dev device.Device, test action.TestStart, timeout time.Duration) error {
	// set up before registering the handler, so responses always see a group
	e.wg = sync.NewExtendedWaitGroup(ctx)
	e.wg.Add(1)
	e.finishOnce = gosync.Once{}

	dev.AddActionHandler(e)
	defer func() {
		dev.RemoveActionHandler(e)
	}()

	// releases the wait group on every exit path, so late responses are no-ops
	defer e.finish()

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

// finish marks the test execution as done. Safe to call multiple times and
// from the device read loop; it never blocks.
func (e *testExecutor) finish() {
	e.finishOnce.Do(e.wg.Done)
}

func (tr *testExecutor) handleExecuteTest(dev device.Device, response *action.Response) {
	if response.Success {
		dev.Log("testrunner", fmt.Sprintf("Timeout: %d", response.Payload.TestDetails.Timeout))
		dev.Log("testrunner", fmt.Sprintf("Categories: %v", response.Payload.TestDetails.Categories))
		if response.Payload.TestDetails.Timeout > 0 {
			tr.wg.UpdateUntil(time.Now().Add(time.Duration(response.Payload.TestDetails.Timeout) * time.Millisecond))
		}
	} else {
		dev.Error("testrunner", "Starting test failed")
		tr.finish()
	}
}

func (tr *testExecutor) handleExecutionResult(dev device.Device, response *action.Response) {
	if response.Success {
		dev.Log("testrunner", "Test finished successfully")
	} else {
		dev.Error("testrunner", "Test finished with errors")
	}
	writer := dev.GetLogWriter()
	if writer != nil {
		writer.Passed(response.Success)
	}
	tr.finish()
}

func (tr *testExecutor) handleTestMethodStart(dev device.Device, response *action.Response) {
	if response.Success {
		writer := dev.GetLogWriter()
		subProtocol, err := tr.protocolWriter.NewSubProtocol(response.Payload.TestDetails.Test, writer)
		if err != nil {
			dev.Error("testrunner", "unable to create new log writer")
			return
		}

		dev.SetLogWriter(subProtocol.Writer)

		dev.Log("testrunner", fmt.Sprintf("Start: %s", response.Payload.TestDetails.Test))
		dev.Log("testrunner", fmt.Sprintf("Timeout: %d", response.Payload.TestDetails.Timeout))
		dev.Log("testrunner", fmt.Sprintf("Categories: %v", response.Payload.TestDetails.Categories))

		timeout := DefaultTestTimeout
		if response.Payload.TestDetails.Timeout > 0 {
			timeout = time.Duration(response.Payload.TestDetails.Timeout) * time.Millisecond
		}

		tr.wg.UpdateUntil(time.Now().Add(timeout))
	}
}

func (tr *testExecutor) handleTestMethodFinished(dev device.Device, response *action.Response) {

	if response.Success {
		dev.Log("testrunner", fmt.Sprintf("Test %s Passed", response.Payload.TestDetails.Test))
	} else {
		dev.Error("testrunner", fmt.Sprintf("Test %s Failed", response.Payload.TestDetails.Test))
	}

	currentWriter := dev.GetLogWriter()
	if currentWriter != nil {
		currentWriter.Passed(response.Success)
		dev.SetLogWriter(currentWriter.Parent())
	}
}

func (tr *testExecutor) handlePerformance(dev device.Device, response *action.Response) {
	dev.LogPerformance(response.Payload.PerformanceData.Checkpoint, response.Payload.PerformanceData.CPU, response.Payload.PerformanceData.FPS, response.Payload.PerformanceData.Memory, response.Payload.PerformanceData.VertexCount, response.Payload.PerformanceData.Triangles, "")
}

func (tr *testExecutor) handleNativeScript(dev device.Device, response *action.Response) {
	go dev.RunNativeScript(*response.Payload.Data)
}

func (tr *testExecutor) handleLog(dev device.Device, response *action.Response) {
	if response.Payload.LogData.Level == action.LogLevel_Exception {
		tr.finish()
	}
}

func (tr *testExecutor) OnActionResponse(d interface{}, response *action.Response) {
	dev := d.(device.Device)
	if response == nil {
		tr.finish()
		dev.Error("testrunner", "Device Disconnected")
		return
	}

	if handler, ok := tr.actionHandler[response.ActionType]; ok {
		handler(dev, response)
	}
}
