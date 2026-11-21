// Package run holds the use cases around test runs: starting one on the
// devices of a test, cancelling it, and reading runs, protocols and their
// logs. Every method enforces the project scope: a run is only reachable
// through its test, the test's app and the app's project.
package run

import (
	"context"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/runs"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
)

// Runner executes one test run; the Unity and the scenario runner are two.
type Runner = tester.Interface

// Key names one run: the test it must live in and its id.
type Key struct {
	testsvc.TestRef
	RunID uint
}

// StartRequest starts the test of the project with the bundle (or the start
// URL, for web and editor apps) on the devices of its configuration.
type StartRequest struct {
	Project  *models.Project
	TestID   uint
	BundleID uint
	StartURL string
	Params   string
}

// RunRef is the run a start created, as the runner stored it.
type RunRef struct {
	Run *models.TestRun
}

// Repo is what the service needs from the storage. A lookup that finds
// nothing returns an apperr.NotFound error.
type Repo interface {
	StartRepo
	ReadRepo
	LookupRepo
}

// Pool is the pool of live devices.
type Pool interface {
	GetDevice(id string) (device.Device, string)
}

// Runners remembers the runner of every run this master started and that is
// still active.
type Runners interface {
	Add(runID uint, r Runner)
	Get(runID uint) (Runner, bool)
	Remove(runID uint)
}

// NewRunner builds the runner for a test configuration type of the project
// and app.
type NewRunner func(kind models.TestType, project *models.Project, appID uint) (Runner, error)

// Summarizer gives each run its status and result counts.
type Summarizer interface {
	Summaries(ctx context.Context, testRuns []models.TestRun) ([]testsvc.RunSummary, error)
}

// Deps are what New needs. RunState and Started are the hooks of the run
// tracker: whether a run executes now, and that one just started.
type Deps struct {
	Repo       Repo
	Pool       Pool
	Runners    Runners
	NewRunner  NewRunner
	Summarizer Summarizer
	RunState   func(runID uint) runs.RunState
	Started    func(runID uint)
}

// Service holds the run use cases.
type Service struct {
	deps Deps
}

// New builds the service; without the tracker hooks no run is running and
// nothing is tracked.
func New(deps Deps) *Service {
	if deps.RunState == nil {
		deps.RunState = func(uint) runs.RunState { return runs.RunUnknown }
	}
	if deps.Started == nil {
		deps.Started = func(uint) {}
	}
	return &Service{deps: deps}
}
