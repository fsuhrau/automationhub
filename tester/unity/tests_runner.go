package unity

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device/node"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	tester_action "github.com/fsuhrau/automationhub/tester/action"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/utils/sync"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultTestTimeout = 5 * time.Minute
)

type workerChannel chan action.TestStart
type cancelChannel chan bool

type testsRunner struct {
	base.TestRunner
	publisher sse.Publisher
	env       map[string]string

	appParams app.Parameter
	projectId string
	appId     uint

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
	if test.TestConfig.Type != models.TestTypeUnity {
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
	// lock devices
	devices := tr.LockDevices(devs)
	if len(devices) == 0 {
		tr.LogError("no lockable devices available")
		return
	}

	defer func() {
		if err := recover(); err != nil {
			tr.LogError("text execution failed recovered: %v", err)
		}
		tr.UnlockDevices(devices)
		tr.TestSessionFinished()
	}()

	tr.LogInfo("Starting devices")
	if err := tr.StartDevices(tr.ctx, devices); err != nil {
		tr.LogError("unable to start devices")
		return
	}

	tr.appParams = base.GetParams(appData, startupUrl)

	// stop app
	tr.LogInfo("Stop apps if running")
	tr.StopApp(tr.ctx, tr.appParams, devices)

	if tr.appParams.App != nil {
		tr.UploadApp(tr.ctx, tr.appParams, devices)

		tr.LogInfo("Install app on devices")
		tr.InstallApp(tr.ctx, tr.appParams, devices)
	}

	tr.LogInfo("Start app on devices and wait for connection")
	connectedDevices, err := tr.StartApp(tr.ctx, tr.appParams, devices, nil, nil)
	if errors.Is(err, sync.TimeoutError) {
		tr.LogError("Timeout while stating app")
	}
	if connectedDevices == nil {
		tr.LogError("No devices connected can't execute tests...")
		return
	}

	tr.LogInfo("Get test list")
	testList, err := tr.getTestList(connectedDevices)
	if err != nil {
		tr.LogError("Get tests failed: %v", err)
	} else {
		if len(testList) == 0 {
			tr.LogInfo("No Tests")
		}
		tr.LogInfo("Execute Tests")
		tr.scheduleTests(connectedDevices, testList)

		tr.LogInfo("Stop Worker")
		tr.cancelFunc()
	}

	tr.LogInfo("Stop apps")
	for _, d := range connectedDevices {
		if err := d.Device.StopApp(&tr.appParams); err != nil {
			tr.LogError("Unable to stop app: %v", err)
		}
	}
}

func (tr *testsRunner) scheduleTests(connectedDevices []base.DeviceMap, testList []models.UnityTestFunction) {
	ctx, cancelFunc := context.WithCancel(tr.ctx)
	defer cancelFunc()
	group := sync.NewExtendedWaitGroup(ctx)

	switch tr.Config.ExecutionType {
	case models.SimultaneouslyExecutionType:

		var workers []workerChannel
		for _, d := range connectedDevices {
			channel := make(workerChannel, len(testList))
			workers = append(workers, channel)
			go tr.workerFunction(ctx, channel, d, group)
		}

		for _, t := range testList {
			a := action.TestStart{
				Assembly: t.Assembly,
				Class:    t.Class,
				Method:   t.Method,
				Env:      tr.env,
			}
			for i := range workers {
				group.Add(1)
				workers[i] <- a
			}
		}

		_ = group.Wait()

		// close worker channel
		for i := range workers {
			close(workers[i])
		}

	case models.ConcurrentExecutionType:
		parallelWorker := make(workerChannel, len(testList))
		for _, d := range connectedDevices {
			go tr.workerFunction(ctx, parallelWorker, d, group)
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

		_ = group.Wait()

		// close worker channel
		close(parallelWorker)
	}
}

func (tr *testsRunner) getTestList(connectedDevices []base.DeviceMap) ([]models.UnityTestFunction, error) {
	var testList []models.UnityTestFunction
	if tr.Config.Unity.UnityTestCategoryType == models.AllTest || tr.Config.Unity.UnityTestCategoryType == models.AllOfCategory {
		tr.LogInfo("UnityTestCategoryType active requesting PlayMode tests")
		actionExecutor := tester_action.NewExecutor(tr.DeviceManager)
		a := &action.TestsGet{}
		if err := actionExecutor.Execute(tr.ctx, connectedDevices[0].Device, a, 5*time.Minute); err != nil {
			return nil, fmt.Errorf("send action to select all tests failed: %v", err)
		}
		testCats := strings.Split(tr.Config.Unity.Categories, ",")
		allCategories := len(testCats) == 0
		for _, t := range a.Tests {
			add := false
			if !allCategories {
				for _, cc := range t.Categories {
					for _, tc := range testCats {
						if tc == cc {
							add = true
							break
						}
					}
				}
			}
			if allCategories || add {
				testList = append(testList, models.UnityTestFunction{
					Assembly: t.Assembly,
					Class:    t.Class,
					Method:   t.Method,
				})
			}
		}
	} else {
		tr.DB.Where("test_config_unity_id = ?", tr.Config.Unity.ID).Find(&testList)
	}
	return testList, nil
}

func (tr *testsRunner) Run(devs []models.Device, binary *models.AppBinary, startupUrl string) (*models.TestRun, error) {
	var params []string
	for k, v := range tr.env {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	var binaryId uint
	if binary != nil {
		binaryId = binary.ID
	}
	if err := tr.InitNewTestSession(binaryId, startupUrl, strings.Join(params, "\n")); err != nil {
		return nil, err
	}

	go tr.exec(devs, binary, startupUrl)

	return &tr.TestRun, nil
}

func (tr *testsRunner) workerFunction(ctx context.Context, channel workerChannel, dev base.DeviceMap, group sync.ExtendedWaitGroup) {
	for {
		select {
		case task := <-channel:
			method := task.Method
			methodParts := strings.Split(method, " ")
			if len(methodParts) > 1 {
				method = methodParts[1]
			}
			tr.runTest(ctx, dev, task, method)
			if !group.IsCanceled() {
				group.Done()
			} else {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (tr *testsRunner) runTest(ctx context.Context, dev base.DeviceMap, task action.TestStart, method string) {

	prot, err := tr.ProtocolWriter.NewProtocol(dev.Model, fmt.Sprintf("%s/%s", task.Class, method))
	if err != nil {
		tr.LogError("Unable to create LogWriter for %s: %v", dev.Device.DeviceID(), err)
	}
	dev.Device.SetLogWriter(prot.Writer)
	defer func() {
		dev.Device.SetLogWriter(nil)
		prot.Close()
	}()
	/*
		tr.LogInfo("Reset device settings")
		var reset action.UnityReset
		resetExec := tester_action.NewExecutor(tr.DeviceManager)
		if err := resetExec.Execute(dev.Device, &reset, 30*time.Second); err != nil {
			tr.LogError("unable to reset settings %s: %v", dev.Device.DeviceID(), err)
		}
	*/
	tr.LogInfo("Run test '%s/%s' on device '%s'", task.Class, method, dev.Device.DeviceID())
	executor := NewExecutor(tr.DeviceManager, tr.ProtocolWriter)
	err = executor.Execute(ctx, dev.Device, task, DefaultTestTimeout)

	passed := prot.Writer.HasPassed()
	finished := "finished successful"
	if !passed {
		finished = "failed"
	}

	if err != nil || len(prot.Errors()) > 0 {
		tr.captureScreenShot(ctx, dev, task, err)
		var errorlist []string
		if err != nil {
			errorlist = append(errorlist, err.Error())
		}
		for _, err := range prot.Errors() {
			errorlist = append(errorlist, err.Error())
		}

		tr.LogError("Test execution %s with errors: %v", finished, strings.Join(errorlist, "\n"))
		return
	}

	tr.LogInfo("Test execution %s ", finished)
}

func (tr *testsRunner) captureScreenShot(ctx context.Context, dev base.DeviceMap, task action.TestStart, err error) {

	nameData := []byte(fmt.Sprintf("%d%s%s%s", time.Now().UnixNano(), tr.TestRun.SessionID, dev.Device.DeviceID(), task.Method))
	fileName := fmt.Sprintf("%x.png", sha1.Sum(nameData))
	filePath := filepath.Join(apps.TestDataPath, fileName)

	rawData, _, _, err := dev.Device.GetScreenshot()
	if rawData == nil {
		var screenshotAction action.GetScreenshot
		actionExecutor := tester_action.NewExecutor(tr.DeviceManager)
		if err := actionExecutor.Execute(ctx, dev.Device, &screenshotAction, 10*time.Second); err != nil {
			tr.LogError("Take screenshot failed: %v", err)
		} else {
			rawData = screenshotAction.ScreenshotData()
		}
	}
	if rawData != nil {
		_ = os.WriteFile(filePath, rawData, os.ModePerm)
		dev.Device.Data("screen", fileName)
	}
}

func (tr *testsRunner) UploadApp(ctx context.Context, params app.Parameter, devices []base.DeviceMap) {

	usedNodes := make(map[manager.NodeIdentifier]manager.Nodes)

	// get unique nodes
	for _, dev := range devices {
		if nodeDev, success := dev.Device.(*node.NodeDevice); success {
			if _, found := usedNodes[nodeDev.GetNodeID()]; !found {
				usedNodes[nodeDev.GetNodeID()] = nodeDev.NodeManager()
			}
		}
	}

	if len(usedNodes) > 0 {
		tr.LogInfo("Upload new App to Nodes")
		for nodeId, mng := range usedNodes {
			err := mng.UploadApp(ctx, nodeId, &params)
			if err != nil {
				tr.LogError("Upload app to nodes failed: %v", err)
			}
		}
	}
}
