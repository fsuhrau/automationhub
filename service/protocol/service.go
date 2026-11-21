// Package protocol holds the use cases around the protocols of a test run:
// the protocol page with the history it is compared with, and the baselines
// a protocol's performance can be compared against. It also owns the
// performance metrics and the run parameter format. Every method enforces
// the project scope: a protocol is only reachable through its run, test,
// app and project.
package protocol

import (
	"context"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Key names one protocol: the test it must live in, its run and its id.
type Key struct {
	testsvc.TestRef
	RunID      uint
	ProtocolID uint
}

// RunScope checks that a run belongs to the test of the scope; it answers
// with a NotFound error when it doesn't.
type RunScope interface {
	CheckRun(ctx context.Context, ref testsvc.TestRef, runID uint) error
}

// Repo is what the service needs from the storage. A lookup that finds
// nothing returns an apperr.NotFound error.
type Repo interface {
	// RunWithProtocol loads the run with the protocol, its device,
	// entries, performance and keys; the run has no protocols when the
	// protocol is not one of its.
	RunWithProtocol(ctx context.Context, runID, protocolID uint) (*models.TestRun, error)
	// History lists the earlier protocols of the same test name and device
	// that measured performance, newest first, among the runs of the app.
	History(ctx context.Context, appID uint, current *models.TestProtocol) ([]models.TestProtocol, error)
	// Protocol loads the protocol of the key if its run, test and app
	// match.
	Protocol(ctx context.Context, key Key) (*models.TestProtocol, error)
	// Candidates lists the protocols the current one can be compared with.
	Candidates(ctx context.Context, q CandidateQuery) ([]models.TestProtocol, error)
	// Baseline loads one candidate with its performance entries.
	Baseline(ctx context.Context, q BaselineQuery) (*models.TestProtocol, error)
	// BundleTags lists the distinct tag lists of the app's bundles.
	BundleTags(ctx context.Context, appID uint) ([]string, error)
}

// Deps are what New needs. DeviceName is how a device is named to a person.
type Deps struct {
	Repo       Repo
	Scope      RunScope
	DeviceName func(*models.Device) string
}

// Service holds the protocol use cases.
type Service struct {
	repo       Repo
	scope      RunScope
	deviceName func(*models.Device) string
}

// New builds the service.
func New(deps Deps) *Service {
	return &Service{repo: deps.Repo, scope: deps.Scope, deviceName: deps.DeviceName}
}

// Page loads the protocol with its run and, when it is there, compares it
// with its history: earlier runs of the same test on the same device.
func (s *Service) Page(ctx context.Context, key Key) (*models.TestRun, error) {
	if err := s.scope.CheckRun(ctx, key.TestRef, key.RunID); err != nil {
		return nil, err
	}
	run, err := s.repo.RunWithProtocol(ctx, key.RunID, key.ProtocolID)
	if err != nil {
		return nil, err
	}
	if len(run.Protocols) == 0 {
		return run, nil
	}
	current := &run.Protocols[0]
	Clean(current)
	history, err := s.repo.History(ctx, key.AppID, current)
	if err != nil {
		return nil, err
	}
	ApplyHistory(current, history)
	return run, nil
}
