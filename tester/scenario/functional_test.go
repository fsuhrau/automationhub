package scenario_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
	"github.com/fsuhrau/automationhub/tester/scenario"
)

func runScenario(t *testing.T, s *basetest.Setup, cfg models.TestConfig, devs []models.Device) (models.TestRun, bool) {
	t.Helper()
	cfg.Type = models.TestTypeScenario
	s.Test.TestConfig = cfg
	runner := scenario.NewWithDeps(s.Deps)
	if err := runner.Initialize(*s.Test, map[string]string{"K": "v"}); err != nil {
		t.Fatal(err)
	}
	wait := basetest.WatchFinished(t)
	run, err := runner.Run(devs, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	return *run, wait(run.ID).Success
}

func logged(s *basetest.Setup, runID uint, text string) bool {
	var n int64
	s.DB.Model(&models.TestRunLogEntry{}).Where("test_run_id = ? AND log LIKE ?", runID, "%"+text+"%").Count(&n)
	return n > 0
}

// The steps run on the started app of each device, the listed tests are
// scheduled, and the run finishes as a success.
func TestScenarioRunsItsStepsAndFinishes(t *testing.T) {
	s := basetest.NewSetup(t, 2)
	s.DB.Create(&models.UnityTestFunction{TestConfigUnityID: 3, Method: "stored"})
	cfg := models.TestConfig{
		Unity: &models.TestConfigUnity{Model: models.Model{ID: 3}, UnityTestCategoryType: models.SelectedTestsOnly},
		Scenario: &models.TestConfigScenario{Steps: []models.ScenarioStep{
			{StepType: models.StepTypeStartApp}, {StepType: models.StepTypeCheckpoint},
		}},
	}

	run, success := runScenario(t, s, cfg, s.Models)

	if !success {
		t.Fatal("run failed")
	}
	for _, id := range []string{"a", "b"} {
		if !logged(s, run.ID, "Scenario step 'checkpoint' on device '"+id+"'") {
			t.Fatalf("steps did not run on device %s", id)
		}
	}
	for _, d := range s.Devices {
		if d.IsLocked() || len(d.CallsTo("StartApp")) != 1 {
			t.Fatalf("device %s: locked %v", d.DeviceID(), d.IsLocked())
		}
	}
}

// A scenario of one step used to crash the run: the step list was indexed
// past its end.
func TestScenarioWithASingleStep(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	cfg := models.TestConfig{Scenario: &models.TestConfigScenario{Steps: []models.ScenarioStep{{StepType: models.StepTypeStopApp}}}}

	run, success := runScenario(t, s, cfg, s.Models)

	if !success || !logged(s, run.ID, "Scenario step 'stop app'") || !logged(s, run.ID, "No Tests") {
		t.Fatalf("success %v", success)
	}
}

// A run without devices fails and finishes.
func TestScenarioWithoutDevicesFailsTheRun(t *testing.T) {
	s := basetest.NewSetup(t, 1)

	run, success := runScenario(t, s, models.TestConfig{}, nil)

	if success || !logged(s, run.ID, "no device available to run the tests") {
		t.Fatalf("success %v", success)
	}
}

func TestScenarioRunnerRejectsAUnityConfig(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	runner := scenario.NewWithDeps(s.Deps)
	s.Test.TestConfig.Type = models.TestTypeUnity
	if err := runner.Initialize(*s.Test, nil); err == nil {
		t.Fatal("unity config accepted")
	}
}

// A runner built from the plain deps of a setup works end to end.
func TestNewWithDepsBuildsARunner(t *testing.T) {
	s := basetest.NewSetup(t, 1)
	runner := scenario.NewWithDeps(s.Deps)
	s.Test.TestConfig = models.TestConfig{Type: models.TestTypeScenario}
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
