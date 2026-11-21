package api

import (
	"io"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// backendOf is the api on an empty database, with the project "proj".
func backendOf(t *testing.T) (*Service, *gorm.DB, *models.Project) {
	t.Helper()
	db := testkit.DB(t)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	svc := New(logger, db, "", &testkit.FakeDeviceManager{}, nil, config.Service{}, &testkit.FakeNodes{})
	return svc, db, testkit.Project(db, func(p *models.Project) { p.Identifier = "proj" })
}

type stubRunner struct{ cancelled string }

func (*stubRunner) Initialize(models.Test, map[string]string) error { return nil }
func (*stubRunner) Run([]models.Device, *models.AppBinary, string) (*models.TestRun, error) {
	return nil, nil
}
func (r *stubRunner) Cancel(id string) error { r.cancelled = id; return nil }
func (*stubRunner) LockedDevices() []uint    { return nil }

func TestRunBookKeepsTheRunnersByRun(t *testing.T) {
	svc, _, _ := backendOf(t)
	book := runBook{svc}
	runner := &stubRunner{}

	book.Add(7, runner)
	got, ok := book.Get(7)
	if !ok || got != runner {
		t.Fatalf("got %v, %v", got, ok)
	}
	if _, ok := book.Get(8); ok {
		t.Fatal("another run has a runner")
	}
	book.Remove(7)
	if _, ok := book.Get(7); ok {
		t.Fatal("the removed runner is still there")
	}
}

func TestAFinishedRunIsForgotten(t *testing.T) {
	svc, _, _ := backendOf(t)
	runBook{svc}.Add(7, &stubRunner{})

	svc.onRunFinished(7)

	if _, ok := (runBook{svc}).Get(7); ok {
		t.Fatal("the finished run's runner is still there")
	}
}

// The run page needs to know whether a run still executes; the run model
// has no status, so it comes from the tracker.
func TestRunStateComesFromTheTracker(t *testing.T) {
	svc, _, _ := backendOf(t)
	svc.runTracker.Started(2)
	svc.runTracker.Started(3)
	svc.runTracker.Finished(3, runs.RunResult{Success: true})

	for run, want := range map[uint]runs.RunState{1: runs.RunUnknown, 2: runs.RunRunning, 3: runs.RunDone} {
		if got := svc.runState(run); got != want {
			t.Errorf("run %d: state %q, want %q", run, got, want)
		}
	}
	// without a tracker (e.g. a service built for other routes) runs aren't running
	svc.runTracker = nil
	if got := svc.runState(2); got != runs.RunUnknown {
		t.Errorf("no tracker: state %q", got)
	}
}

func TestNewRunnerBuildsTheRunnerOfTheTestType(t *testing.T) {
	svc, _, project := backendOf(t)
	for _, kind := range []models.TestType{models.TestTypeUnity, models.TestTypeScenario} {
		runner, err := svc.newRunner(kind, project, 1)
		if err != nil || runner == nil {
			t.Errorf("type %v: %v, %v", kind, runner, err)
		}
	}
	if _, err := svc.newRunner(models.TestType(99), project, 1); apperr.KindOf(err) != apperr.KindInternal {
		t.Errorf("unknown type: %v", err)
	}
}
