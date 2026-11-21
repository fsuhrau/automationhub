// Package overview holds the project overview and the project statistics:
// what a project has, which tests need attention, what runs now and what
// ran last. Every method is scoped to one project.
package overview

import (
	"context"

	"github.com/fsuhrau/automationhub/device"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Totals is what a project has: apps, bundles and the bytes of the bundles
// that still have their file.
type Totals struct {
	Apps          int64
	Bundles       int64
	BundleStorage int64
}

// ProtocolQuery asks for the newest protocols of a project, only the failed
// ones when FailedOnly is set.
type ProtocolQuery struct {
	ProjectID  uint
	FailedOnly bool
	Limit      int
}

// Repo is what the service needs from the storage.
type Repo interface {
	Totals(ctx context.Context, projectID uint) (Totals, error)
	// Tests lists the tests of the project's apps, by id.
	Tests(ctx context.Context, projectID uint) ([]models.Test, error)
	// LastRuns lists the newest run of each test, newest first.
	LastRuns(ctx context.Context, testIDs []uint) ([]models.TestRun, error)
	// LatestRuns lists the newest runs of the tests, newest first.
	LatestRuns(ctx context.Context, testIDs []uint, limit int) ([]models.TestRun, error)
	// Protocols lists the newest protocols with their run, test and device.
	Protocols(ctx context.Context, q ProtocolQuery) ([]models.TestProtocol, error)
	// Devices lists the devices the project can see.
	Devices(ctx context.Context, project *models.Project) ([]models.Device, error)
}

// Summarizer gives each run its status and result counts.
type Summarizer interface {
	Summaries(ctx context.Context, testRuns []models.TestRun) ([]testsvc.RunSummary, error)
}

// Pool is the pool of live devices.
type Pool interface {
	GetDevice(id string) (device.Device, string)
}

// Deps are what New needs.
type Deps struct {
	Repo       Repo
	Summarizer Summarizer
	Pool       Pool
}

// Service holds the overview use cases.
type Service struct {
	deps Deps
}

// New builds the service.
func New(deps Deps) *Service {
	return &Service{deps: deps}
}
