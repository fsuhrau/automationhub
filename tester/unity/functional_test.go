package unity_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
	"github.com/fsuhrau/automationhub/tester/unity"
)

// runUnity runs a unity test through the public entry point and returns the
// finished event and the protocols the run wrote.
func runUnity(t *testing.T, s *basetest.Setup, cfg models.TestConfig, binary *models.AppBinary) (models.TestRun, bool, []models.TestProtocol) {
	t.Helper()
	cfg.Type = models.TestTypeUnity
	s.Test.TestConfig = cfg
	runner := unity.NewWithDeps(s.Deps)
	if err := runner.Initialize(*s.Test, map[string]string{"LEVEL": "3"}); err != nil {
		t.Fatal(err)
	}
	wait := basetest.WatchFinished(t)
	run, err := runner.Run(s.Models, binary, "")
	if err != nil {
		t.Fatal(err)
	}
	finished := wait(run.ID)
	var prots []models.TestProtocol
	s.DB.Where("test_run_id = ?", run.ID).Find(&prots)
	return *run, finished.Success, prots
}

// The app lists its tests; every test runs once on the free devices, and the
// run finishes as a success with a protocol per test.
func TestUnityRunsTheTestsTheAppLists(t *testing.T) {
	s := basetest.NewSetup(t, 2)
	s.Manager.Tests = []action.Test{{Class: "A", Method: "one"}, {Class: "B", Method: "two"}, {Class: "C", Method: "three"}}
	cfg := models.TestConfig{ExecutionType: models.ConcurrentExecutionType, Unity: &models.TestConfigUnity{UnityTestCategoryType: models.AllTest}}

	run, success, prots := runUnity(t, s, cfg, &models.AppBinary{Identifier: "com.example", AppPath: "x.apk"})

	if !success || len(prots) != 3 {
		t.Fatalf("success %v, %d protocols, want 3", success, len(prots))
	}
	if run.Parameter != "LEVEL=3" {
		t.Fatalf("run parameter %q", run.Parameter)
	}
	for _, d := range s.Devices {
		if d.IsLocked() || len(d.CallsTo("StartApp")) != 1 || len(d.CallsTo("InstallApp")) != 0 {
			t.Fatalf("device %s: locked %v", d.DeviceID(), d.IsLocked())
		}
	}
	for _, c := range s.Manager.CallsTo("SendAction") {
		if start, ok := c.Args[1].(*action.TestStart); ok && start.Env["LEVEL"] != "3" {
			t.Fatalf("test start without env: %+v", start)
		}
	}
}

// The tests of the config run on every device when they run simultaneously;
// a failing test fails the run.
func TestUnityRunsTheConfiguredTestsOnEveryDevice(t *testing.T) {
	s := basetest.NewSetup(t, 2)
	s.Manager.TestPassed = false
	s.DB.Create(&models.UnityTestFunction{TestConfigUnityID: 9, Class: "A", Method: "one"})
	s.DB.Create(&models.UnityTestFunction{TestConfigUnityID: 9, Class: "B", Method: "two"})
	cfg := models.TestConfig{
		ExecutionType: models.SimultaneouslyExecutionType,
		Unity:         &models.TestConfigUnity{Model: models.Model{ID: 9}, UnityTestCategoryType: models.SelectedTestsOnly},
	}

	_, success, prots := runUnity(t, s, cfg, nil)

	if success || len(prots) != 4 {
		t.Fatalf("success %v, %d protocols, want 4", success, len(prots))
	}
	for _, p := range prots {
		if p.TestResult != models.TestResultFailed {
			t.Fatalf("protocol %s has result %v", p.TestName, p.TestResult)
		}
	}
}

// A config without unity settings can't list tests: the run fails, and the
// devices are released.
func TestUnityFailsWithoutUnitySettings(t *testing.T) {
	s := basetest.NewSetup(t, 1)

	_, success, prots := runUnity(t, s, models.TestConfig{}, nil)

	if success || len(prots) != 0 || s.Devices[0].IsLocked() {
		t.Fatalf("success %v, %d protocols, locked %v", success, len(prots), s.Devices[0].IsLocked())
	}
}

func TestUnityRunnerRejectsAScenarioConfig(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	runner := unity.NewWithDeps(s.Deps)
	s.Test.TestConfig.Type = models.TestTypeScenario
	if err := runner.Initialize(*s.Test, nil); err == nil {
		t.Fatal("scenario config accepted")
	}
}

// The deprecated constructor still builds a working runner.
func TestDeprecatedNewBuildsARunner(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	runner := unity.New(s.DB, "node", s.Manager, nil, "p", 1)
	s.Test.TestConfig = models.TestConfig{Type: models.TestTypeUnity, Unity: &models.TestConfigUnity{UnityTestCategoryType: models.AllTest}}
	if err := runner.Initialize(*s.Test, nil); err != nil {
		t.Fatal(err)
	}
	wait := basetest.WatchFinished(t)
	run, err := runner.Run(s.Models, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !wait(run.ID).Success {
		t.Fatal("run failed")
	}
}
