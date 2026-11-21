package wiring

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func testInfra(t *testing.T) Infra {
	t.Helper()
	return Infra{
		DB: testkit.DB(t), Devices: &testkit.FakeDeviceManager{}, Nodes: &testkit.FakeNodes{},
		Config: config.Service{NodeUrl: "http://node", Paths: config.Paths{AppStore: t.TempDir(), TestData: t.TempDir()}},
		Logger: quietLoggerInternal(),
	}
}

func TestRunnerFactoryBuildsTheRunnerOfTheTestType(t *testing.T) {
	factory := runnerFactory(testInfra(t))
	project := &models.Project{Identifier: "proj"}

	for _, kind := range []models.TestType{models.TestTypeUnity, models.TestTypeScenario} {
		runner, err := factory(kind, project, 1)
		if err != nil || runner == nil {
			t.Errorf("type %v: %v, %v", kind, runner, err)
		}
	}
	if _, err := factory(models.TestType(99), project, 1); apperr.KindOf(err) != apperr.KindInternal {
		t.Errorf("unknown type: %v", err)
	}
}

// The run page needs to know whether a run still executes; the run model
// has no status, so it comes from the tracker.
func TestRunStateComesFromTheTracker(t *testing.T) {
	rt := newRuntime(&testkit.FakeDeviceManager{})
	rt.tracker.Started(2)
	rt.tracker.Started(3)
	rt.tracker.Finished(3, runs.RunResult{Success: true})

	for run, want := range map[uint]runs.RunState{1: runs.RunUnknown, 2: runs.RunRunning, 3: runs.RunDone} {
		if got := rt.runState(run); got != want {
			t.Errorf("run %d: state %q, want %q", run, got, want)
		}
	}
}

func TestAFinishedRunIsForgottenByTheRegistry(t *testing.T) {
	rt := newRuntime(&testkit.FakeDeviceManager{})
	rt.runners.Add(7, nil)
	rt.tracker.Started(7)

	rt.tracker.Finished(7, runs.RunResult{})

	if _, ok := rt.runners.Get(7); ok {
		t.Fatal("the finished run's runner is still there")
	}
}
