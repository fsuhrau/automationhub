// Package test is the use cases of the test definitions: listing and
// reading the tests of an app, creating them and saving the edit page. Every
// method enforces the project scope: a test is only reachable through its
// app and the app's project.
package test

import (
	"context"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Scope is where a test lives: the app and the project that owns it.
type Scope struct {
	ProjectID uint
	AppID     uint
}

// TestRef names one test: the scope it must live in and its id.
type TestRef struct {
	Scope
	TestID uint
}

// ResultCount is how many test protocols of a run ended with a result.
type ResultCount struct {
	TestRunID  uint
	TestResult models.TestResultState
	N          int
}

// LogErrorCount is how many error entries the log of a run has.
type LogErrorCount struct {
	TestRunID uint
	N         int
}

// Repo is what the service needs from the storage. Lookups that find
// nothing return an apperr.NotFound error.
type Repo interface {
	// Tests lists the tests of the scope with their config, ordered by id.
	Tests(ctx context.Context, sc Scope) ([]models.Test, error)
	// Test loads one test of the scope with its config in detail.
	Test(ctx context.Context, ref TestRef) (*models.Test, error)
	// AppExists tells whether the app belongs to the project.
	AppExists(ctx context.Context, sc Scope) (bool, error)
	// CountDevices counts how many of the devices the project can see.
	CountDevices(ctx context.Context, projectID uint, ids []uint) (int, error)
	// CreateTest stores the test with its config in one transaction and
	// returns the test row.
	CreateTest(ctx context.Context, test models.Test) (models.Test, error)
	// UpdateTest saves the desired config over the loaded test in one
	// transaction: devices and Unity test functions are replaced.
	UpdateTest(ctx context.Context, current *models.Test, desired models.Test) error
	// RecentRuns lists the newest runs of a test, newest first.
	RecentRuns(ctx context.Context, testID uint, limit int) ([]models.TestRun, error)
	// RunResults counts the test protocols of the runs per result.
	RunResults(ctx context.Context, runIDs []uint) ([]ResultCount, error)
	// RunLogErrors counts the error log entries of the runs.
	RunLogErrors(ctx context.Context, runIDs []uint) ([]LogErrorCount, error)
}

// RunStateFunc tells whether this master is running a run right now.
type RunStateFunc func(runID uint) runs.RunState

// Deps are what New needs.
type Deps struct {
	Repo     Repo
	RunState RunStateFunc
}

// Service holds the test definition use cases.
type Service struct {
	repo     Repo
	runState RunStateFunc
}

// New builds the service; without a RunState every run is not running.
func New(deps Deps) *Service {
	state := deps.RunState
	if state == nil {
		state = func(uint) runs.RunState { return runs.RunUnknown }
	}
	return &Service{repo: deps.Repo, runState: state}
}
