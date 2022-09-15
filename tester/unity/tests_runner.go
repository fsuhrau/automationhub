package unity

import (
	"crypto/sha1"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	tester_action "github.com/fsuhrau/automationhub/tester/action"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/utils/sync"
	"gorm.io/gorm"
	"net"
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
	fin       chan bool
	publisher sse.Publisher
	env       map[string]string

	appParams app.Parameter
}

func New(db *gorm.DB, ip net.IP, deviceManager manager.Devices, publisher sse.Publisher) *testsRunner {
	testRunner := &testsRunner{
		fin:       make(chan bool, 1),
		publisher: publisher,
	}
	testRunner.Init(deviceManager, ip, db)
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

func (tr *testsRunner) exec(devs []models.Device, appData *models.AppBinary) {
	defer tr.TestSessionFinished()

	// lock devices
	devices := tr.LockDevices(devs)
	if len(devices) == 0 {
		tr.LogError("no lockable devices available")
		return
	}
	defer tr.UnlockDevices(devices)

	tr.LogInfo("Starting devices")
	if err := tr.StartDevices(devices); err != nil {
		tr.LogError("unable to start devices")
		return
	}

	if appData != nil {
		tr.appParams = app.Parameter{
			AppBinaryID:    appData.ID,
			Identifier:     appData.App.Identifier,
			AppPath:        appData.AppPath,
			LaunchActivity: appData.LaunchActivity,
			Name:           appData.Name,
			Version:        appData.Version,
			Hash:           appData.Hash,
		}
	} else {
		tr.appParams = app.Parameter{
			AppBinaryID:    appData.ID,
			LaunchActivity: "BootScene",
		}
	}

	// stop app
	tr.LogInfo("stop apps if running")
	tr.StopApp(tr.appParams, devices)

	tr.LogInfo("install app on devices")
	tr.InstallApp(tr.appParams, devices)

	tr.LogInfo("start app on devices and wait for connection")
	connectedDevices, err := tr.StartApp(tr.appParams, devices, nil, nil)
	if err == sync.TimeoutError {
		tr.LogError("one or more apps didn't connect")
	}
	if connectedDevices == nil {
		tr.LogError("no devices connected can't execute tests...")
		return
	}

	testList, err := tr.getTestList(connectedDevices)
	if err != nil {
		tr.LogError("get tests failed: %v", err)
	} else {
		tr.LogInfo("Execute Tests")
		tr.scheduleTests(connectedDevices, testList)
	}

	tr.LogInfo("stop apps")
	for _, d := range connectedDevices {
		if err := d.Device.StopApp(&tr.appParams); err != nil {
			tr.LogError("unable to stop app: %v", err)
		}
	}
}

func (tr *testsRunner) scheduleTests(connectedDevices []base.DeviceMap, testList []models.UnityTestFunction) {
	switch tr.Config.ExecutionType {
	case models.SimultaneouslyExecutionType:
		// each device gets its own input pool which needs to be processed
		cancel := make(cancelChannel, len(connectedDevices))
		group := sync.ExtendedWaitGroup{}

		var workers []workerChannel
		for _, d := range connectedDevices {
			channel := make(workerChannel, len(testList))
			workers = append(workers, channel)
			go tr.WorkerFunction(channel, d, cancel, &group)
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
		// have one input pool where each device can select a job
		cancel := make(cancelChannel, len(connectedDevices))
		group := sync.ExtendedWaitGroup{}

		parallelWorker := make(workerChannel, len(testList))
		for _, d := range connectedDevices {
			go tr.WorkerFunction(parallelWorker, d, cancel, &group)
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
		for i := 0; i < len(connectedDevices); i++ {
			cancel <- true
		}
		close(cancel)
		close(parallelWorker)
	}
}

func (tr *testsRunner) getTestList(connectedDevices []base.DeviceMap) ([]models.UnityTestFunction, error) {
	var testList []models.UnityTestFunction
	if tr.Config.Unity.RunAllTests {
		tr.LogInfo("RunAllTests active requesting PlayMode tests")
		actionExecutor := tester_action.NewExecutor(tr.DeviceManager)
		a := &action.TestsGet{}
		if err := actionExecutor.Execute(connectedDevices[0].Device, a, 5*time.Minute); err != nil {

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

func (tr *testsRunner) Run(devs []models.Device, appData *models.AppBinary) (*models.TestRun, error) {
	var params []string
	for k, v := range tr.env {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	var appID uint
	if appData != nil {
		appID = appData.ID
	}
	if err := tr.InitNewTestSession(appID, strings.Join(params, "\n")); err != nil {
		return nil, err
	}

	go tr.exec(devs, appData)

	return &tr.TestRun, nil
}

func (tr *testsRunner) WorkerFunction(channel workerChannel, dev base.DeviceMap, cancel cancelChannel, group *sync.ExtendedWaitGroup) {
	for {
		select {
		case task := <-channel:
			method := task.Method
			methodParts := strings.Split(method, " ")
			if len(methodParts) > 1 {
				method = methodParts[1]
			}
			tr.runTest(dev, task, method)
			group.Done()
		case <-cancel:
			return
		}
	}
}

func (tr *testsRunner) runTest(dev base.DeviceMap, task action.TestStart, method string) {
	prot, err := tr.ProtocolWriter.NewProtocol(dev.Model, fmt.Sprintf("%s/%s", task.Class, method))
	if err != nil {
		tr.LogError("unable to create LogWriter for %s: %v", dev.Device.DeviceID(), err)
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
	executor := NewExecutor(tr.DeviceManager)
	err = executor.Execute(dev.Device, task, DefaultTestTimeout)
	if err != nil || len(prot.Errors()) > 0 {
		nameData := []byte(fmt.Sprintf("%d%s%s%s", time.Now().UnixNano(), tr.TestRun.SessionID, dev.Device.DeviceID(), task.Method))
		filePath := fmt.Sprintf("test/data/%x.png", sha1.Sum(nameData))

		rawData, _, _, err := dev.Device.GetScreenshot()
		if rawData == nil {
			var screenshotAction action.GetScreenshot
			actionExecutor := tester_action.NewExecutor(tr.DeviceManager)
			if err := actionExecutor.Execute(dev.Device, &screenshotAction, 10*time.Second); err != nil {
				tr.LogError("take screenshot failed: %v", err)
			} else {
				rawData = screenshotAction.ScreenshotData()
			}
		}
		if rawData != nil {
			dir, _ := filepath.Split(filePath)
			os.MkdirAll(dir, os.ModePerm)
			os.WriteFile(filePath, rawData, os.ModePerm)
			dev.Device.Data("screen", filePath)
		}
		tr.LogError("test execution failed: %v", err)
	} else {
		tr.LogInfo("test execution finished")
	}
}
