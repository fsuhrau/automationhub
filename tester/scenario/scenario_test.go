package scenario

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
)

type fixture struct {
	*basetest.Setup
	runner   *base.Runner
	strategy *strategy
}

func newFixture(t *testing.T, cfg models.TestConfig) *fixture {
	t.Helper()
	f := &fixture{Setup: basetest.NewSetup(t, 1)}
	f.runner = base.NewRunner(f.Deps, models.TestTypeScenario, func(r *base.Runner) base.Strategy {
		f.strategy = &strategy{runner: r}
		return f.strategy
	})
	cfg.Type = models.TestTypeScenario
	f.Test.TestConfig = cfg
	if err := f.runner.Initialize(*f.Test, nil); err != nil {
		t.Fatal(err)
	}
	if err := f.runner.InitNewTestSession(context.Background(), base.Session{}); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) logs(level string) []string {
	var entries []models.TestRunLogEntry
	f.DB.Where("test_run_id = ? AND level = ?", f.runner.TestRun.ID, level).Order("id").Find(&entries)
	var out []string
	for _, e := range entries {
		out = append(out, e.Log)
	}
	return out
}

func scenarioWith(types ...models.StepType) models.TestConfig {
	var steps []models.ScenarioStep
	for _, typ := range types {
		steps = append(steps, models.ScenarioStep{StepType: typ})
	}
	return models.TestConfig{Scenario: &models.TestConfigScenario{Steps: steps}}
}

func TestAppStartedRunsTheStepsInOrder(t *testing.T) {
	f := newFixture(t, scenarioWith(models.StepTypeInstallApp, models.StepTypeCheckpoint, models.StepTypeExecuteTest))
	f.strategy.AppStarted(context.Background(), f.Devices[0])

	want := []string{
		"Scenario step 'install app' on device 'a'",
		"Scenario step 'checkpoint' on device 'a'",
		"Scenario step 'execute test' on device 'a'",
	}
	got := f.logs("log")
	if len(got) != 3 {
		t.Fatalf("logs %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("log %d: %q, want %q", i, got[i], want[i])
		}
	}
}

// A scenario without steps never panics;
// the step list used to be indexed past its end.
func TestAppStartedWithoutStepsDoesNothing(t *testing.T) {
	f := newFixture(t, models.TestConfig{})
	f.strategy.AppStarted(context.Background(), f.Devices[0])
	if len(f.logs("log")) != 0 {
		t.Fatal("logged without steps")
	}
}

func TestAppStartedStopsWhenTheRunIsCancelled(t *testing.T) {
	f := newFixture(t, scenarioWith(models.StepTypeStartApp, models.StepTypeStopApp))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f.strategy.AppStarted(ctx, f.Devices[0])
	if len(f.logs("log")) != 0 {
		t.Fatal("step ran after the cancel")
	}
}

func TestStepNames(t *testing.T) {
	tests := map[models.StepType]string{
		models.StepTypeInstallApp:   "install app",
		models.StepTypeUninstallApp: "uninstall app",
		models.StepTypeStartApp:     "start app",
		models.StepTypeStopApp:      "stop app",
		models.StepTypeCheckpoint:   "checkpoint",
		models.StepTypeExecuteTest:  "execute test",
		models.StepType(99):         "unknown",
	}
	for typ, want := range tests {
		if got := stepName(typ); got != want {
			t.Errorf("stepName(%d) = %q, want %q", typ, got, want)
		}
	}
}

func TestTestListOfAScenario(t *testing.T) {
	cfg := scenarioWith()
	f := newFixture(t, cfg)
	connected := []base.DeviceMap{{Device: f.Devices[0]}}

	got, err := f.strategy.TestList(context.Background(), connected)
	if err != nil || got != nil {
		t.Fatalf("scenario without unity settings: %v, %v", got, err)
	}

	f.runner.Config.Unity = &models.TestConfigUnity{Model: models.Model{ID: 4}, UnityTestCategoryType: models.SelectedTestsOnly}
	f.DB.Create(&models.UnityTestFunction{TestConfigUnityID: 4, Method: "stored"})
	got, err = f.strategy.TestList(context.Background(), connected)
	if err != nil || len(got) != 1 || got[0].Method != "stored" {
		t.Fatalf("scenario with unity settings: %v, %v", got, err)
	}
}

func TestRunTestDoesNothingYet(t *testing.T) {
	f := newFixture(t, models.TestConfig{})
	f.strategy.RunTest(context.Background(), base.DeviceMap{Device: f.Devices[0]}, action.TestStart{})
	if len(f.Manager.Calls()) != 0 || len(f.Devices[0].Calls()) != 0 {
		t.Fatal("scenario test execution is not implemented, nothing may happen")
	}
}
