package scenario

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/utils/sync"
	"gorm.io/gorm"
	"strings"
)

type workerChannel chan action.TestStart
type cancelChannel chan bool

type testsRunner struct {
	base.TestRunner
	publisher sse.Publisher
	env       map[string]string

	appParams app.Parameter

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func New(db *gorm.DB, nodeUrl string, deviceManager manager.Devices, publisher sse.Publisher, projectId string, appId uint) *testsRunner {
	ctx, cancelFunc := context.WithCancel(context.Background())
	testRunner := &testsRunner{
		publisher:  publisher,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
	testRunner.Init(deviceManager, nodeUrl, db, projectId, appId)
	return testRunner
}

func (tr *testsRunner) Initialize(test models.Test, env map[string]string) error {
	if test.TestConfig.Type != models.TestTypeScenario {
		return fmt.Errorf("config needs to be unity to create a unity test handler")
	}
	tr.env = env
	tr.Test = test
	tr.Config = test.TestConfig
	return nil
}

func (tr *testsRunner) Cancel(runId string) error {
	tr.cancelFunc()
	return nil
}

func (tr *testsRunner) exec(devs []models.Device, appData *models.AppBinary, startupUrl string) {
	defer tr.TestSessionFinished()

	// lock devices
	devices := tr.LockDevices(devs)
	if len(devices) == 0 {
		tr.LogError("No lockable devices available")
		return
	}
	defer tr.UnlockDevices(devices)

	tr.LogInfo("Starting devices")
	if err := tr.StartDevices(tr.ctx, devices); err != nil {
		tr.LogError("Unable to start devices")
		return
	}

	tr.appParams = base.GetParams(appData, startupUrl)

	// stop app
	tr.LogInfo("Stop apps if running")
	tr.StopApp(tr.ctx, tr.appParams, devices)

	tr.LogInfo("Install app on devices")
	tr.InstallApp(tr.ctx, tr.appParams, devices)

	tr.LogInfo("Start app on devices and wait for connection")
	connectedDevices, err := tr.StartApp(tr.ctx, tr.appParams, devices, func(d device.Device) {
		go tr.executeSequence(d, 0, tr.Config.Scenario.Steps)
	}, nil)
	if err == sync.TimeoutError {
		tr.LogError("Timeout while stating app")
	}

	var testList []models.UnityTestFunction
	if tr.Config.Unity.UnityTestCategoryType == models.AllTest {
		tr.LogInfo("UnityTestCategoryType active requesting PlayMode tests")
		a := &action.TestsGet{}
		if err := tr.DeviceManager.SendAction(connectedDevices[0].Device, a); err != nil {
			tr.LogError("Send action to select all tests failed: %v", err)
			return
		}
		for _, t := range a.Tests {
			testList = append(testList, models.UnityTestFunction{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
			})
		}
	} else {
		tr.DB.Where("test_config_unity_id = ?", tr.Config.Unity.ID).Find(&testList)
	}

	tr.LogInfo("Execute Tests")

	group := sync.NewExtendedWaitGroup(tr.ctx)

	switch tr.Config.ExecutionType {
	case models.SimultaneouslyExecutionType:
		// each device gets its own input pool which needs to be processed
		cancel := make(cancelChannel, len(connectedDevices))

		var workers []workerChannel
		for _, d := range connectedDevices {
			channel := make(workerChannel, len(testList))
			workers = append(workers, channel)
			go tr.workerFunction(channel, d, group)
		}

		for _, t := range testList {
			a := action.TestStart{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
			}
			for i := range workers {
				group.Add(1)
				workers[i] <- a
			}
		}
		group.Wait()
		for i := 0; i < len(connectedDevices); i++ {
			cancel <- true
		}
		close(cancel)
		for i := range workers {
			close(workers[i])
		}

	case models.ConcurrentExecutionType:

		parallelWorker := make(workerChannel, len(testList))
		for _, d := range connectedDevices {
			go tr.workerFunction(parallelWorker, d, group)
		}
		for _, t := range testList {
			a := action.TestStart{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
				Env:      tr.env,
			}
			group.Add(1)
			parallelWorker <- a
		}
		group.Wait()
		close(parallelWorker)
	}

	tr.LogInfo("Stop apps")
	for _, d := range connectedDevices {
		if err := d.Device.StopApp(&tr.appParams); err != nil {
			tr.LogError("Unable to start app: %v", err)
		}
	}
}

func (tr *testsRunner) OnDeviceConnected(d device.Device) {

}

func (tr *testsRunner) Run(devs []models.Device, binary *models.AppBinary, startURL string) (*models.TestRun, error) {
	var params []string
	for k, v := range tr.env {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}
	var binaryID uint
	if binary != nil {
		binaryID = binary.ID
	}
	if err := tr.InitNewTestSession(binaryID, startURL, strings.Join(params, "\n")); err != nil {
		return nil, err
	}

	go tr.exec(devs, binary, startURL)

	return &tr.TestRun, nil
}

func (tr *testsRunner) workerFunction(channel workerChannel, dev base.DeviceMap, group sync.ExtendedWaitGroup) {
	for {
		select {
		case task, ok := <-channel:
			if !ok {
				return
			}
			method := task.Method
			methodParts := strings.Split(method, " ")
			if len(methodParts) > 1 {
				method = methodParts[1]
			}
			tr.runTest(dev, task, method)
			group.Done()
		case <-tr.ctx.Done():
			tr.LogInfo("Test run cancelled by context")
			return
		}
	}
}

func (tr *testsRunner) runTest(dev base.DeviceMap, task action.TestStart, method string) {
	/*
		prot, err := tr.ProtocolWriter.NewProtocol(dev.Model.id, fmt.Sprintf("%s/%s", task.class, method))
		if err != nil {
			tr.LogError("unable to create LogWriter for %s: %v", dev.device.deviceId(), err)
		}
		dev.device.SetLogWriter(prot.Writer)
		defer func() {
			dev.device.SetLogWriter(nil)
			prot.Close()
		}()

		tr.LogInfo("Run test '%s/%s' on device '%s'", task.class, method, dev.device.deviceId())
		executor := NewExecutor(tr.DeviceManager)
		if err := executor.Execute(dev.device, task, 5*time.Minute); err != nil {
			rawData, _, _, err := dev.device.GetScreenshot()
			nameData := []byte(fmt.Sprintf("%d%s%s%s", time.Now().UnixNano(), tr.testRun.sessionId, dev.device.deviceId(), task.method))
			filePath := fmt.Sprintf("test/data/%x.png", sha1.Sum(nameData))
			dir, _ := filepath.Split(filePath)
			os.MkdirAll(dir, os.ModePerm)
			os.WriteFile(filePath, rawData, os.ModePerm)
			dev.device.data("screen", filePath)
			tr.LogError("test execution failed: %v", err)
		} else {
			tr.LogInfo("test execution finished")
		}
	*/
}

func (tr *testsRunner) executeSequence(d device.Device, index int, steps []models.ScenarioStep) {

	currentStep := steps[index]

	switch currentStep.StepType {
	case models.StepTypeInstallApp:
		tr.stepInstallApp(d, index, steps)
	case models.StepTypeUninstallApp:
		tr.stepUninstallApp(d, index, steps)
	case models.StepTypeStartApp:
		tr.stepStartApp(d, index, steps)
	case models.StepTypeStopApp:
		tr.stepStopApp(d, index, steps)
	case models.StepTypeCheckpoint:
		tr.stepCheckpoint(d, index, steps)
	case models.StepTypeExecuteTest:
		tr.stepExecuteTest(d, index, steps)
	}
}

func (tr *testsRunner) stepInstallApp(d device.Device, index int, steps []models.ScenarioStep) {

	//	currentStep := steps[index]
	//	currentStep.AppIdentifier
	//	app.parameter{}
	//	d.InstallApp()

	tr.executeSequence(d, index+1, steps)
}

func (tr *testsRunner) stepUninstallApp(d device.Device, index int, steps []models.ScenarioStep) {

	tr.executeSequence(d, index+1, steps)
}

func (tr *testsRunner) stepStartApp(d device.Device, index int, steps []models.ScenarioStep) {

	tr.executeSequence(d, index+1, steps)
}

func (tr *testsRunner) stepStopApp(d device.Device, index int, steps []models.ScenarioStep) {

	tr.executeSequence(d, index+1, steps)
}

func (tr *testsRunner) stepCheckpoint(d device.Device, index int, steps []models.ScenarioStep) {

	tr.executeSequence(d, index+1, steps)
}

func (tr *testsRunner) stepExecuteTest(d device.Device, index int, steps []models.ScenarioStep) {

	tr.executeSequence(d, index+1, steps)
}
