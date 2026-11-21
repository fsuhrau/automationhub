package run

import (
	"context"
	"time"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// DeviceQuery selects devices the project can see: all of them, or only
// those with the ids.
type DeviceQuery struct {
	Project *models.Project
	All     bool
	IDs     []uint
}

// StartRepo is what starting and cancelling a run needs.
type StartRepo interface {
	// StartTest loads a test of the project with its app, config, devices
	// and Unity settings.
	StartTest(ctx context.Context, projectID, testID uint) (*models.Test, error)
	// Bundle loads a bundle of the app, with the app.
	Bundle(ctx context.Context, appID, bundleID uint) (*models.AppBinary, error)
	// Devices lists the devices of the query with their node, by id.
	Devices(ctx context.Context, q DeviceQuery) ([]models.Device, error)
	// RunInProject tells that the run belongs to the project.
	RunInProject(ctx context.Context, projectID, runID uint) error
}

// PageQuery asks for at most Limit runs of a test, older than the run id
// Before when it is not 0.
type PageQuery struct {
	TestID uint
	Before uint
	Limit  int
}

// StartupQuery asks for the startup times of the device of Status in the
// earlier runs of the app, at most Limit runs back.
type StartupQuery struct {
	AppID  uint
	Status *models.TestRunDeviceStatus
	Limit  int
}

// Neighbours are the runs of a test before and after a run; 0 for none.
type Neighbours struct {
	Prev uint
	Next uint
}

// RunEnd is the end of one protocol of a run.
type RunEnd struct {
	TestRunID uint
	EndedAt   time.Time
}

// ReadRepo is what reading runs needs.
type ReadRepo interface {
	// TestInScope tells that the test belongs to the app and the project.
	TestInScope(ctx context.Context, ref testsvc.TestRef) error
	// RunInTest tells that the run belongs to the test.
	RunInTest(ctx context.Context, testID, runID uint) error
	// TestRuns pages the runs of a test, newest first.
	TestRuns(ctx context.Context, q PageQuery) ([]models.TestRun, error)
	// RunEnds lists the end of every ended protocol of the runs.
	RunEnds(ctx context.Context, runIDs []uint) ([]RunEnd, error)
	// LastRun loads the newest run of a test in detail.
	LastRun(ctx context.Context, testID uint) (*models.TestRun, error)
	// Run loads one run in detail, with its recordings.
	Run(ctx context.Context, runID uint) (*models.TestRun, error)
	// Neighbours finds the runs of the test before and after a run.
	Neighbours(ctx context.Context, testID, runID uint) (Neighbours, error)
	// StartupTimes lists the startup times of the device of a status in the
	// earlier runs of the app.
	StartupTimes(ctx context.Context, q StartupQuery) ([]models.TestRunDeviceStatus, error)
}

// LogQuery asks for the log entries of a protocol of the project: those of
// the level, containing the text, paged by Offset and Limit.
type LogQuery struct {
	ProjectID  uint
	ProtocolID uint
	Level      string
	Contains   string
	Offset     int
	Limit      int
}

// LookupRepo is what the device and protocol lookups need.
type LookupRepo interface {
	// Device loads a device the project can see, with its node.
	Device(ctx context.Context, project *models.Project, id uint) (*models.Device, error)
	// ProjectProtocol loads a protocol of the project with its device and
	// keys.
	ProjectProtocol(ctx context.Context, projectID, protocolID uint) (*models.TestProtocol, error)
	// ProtocolLogs pages the log entries of the query and counts all that
	// match.
	ProtocolLogs(ctx context.Context, q LogQuery) ([]models.ProtocolEntry, int64, error)
}
