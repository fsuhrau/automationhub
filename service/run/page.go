package run

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/service/protocol"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Page is the run page: the run in detail and its neighbours.
type Page struct {
	TestRun   models.TestRun `json:"testRun"`
	PrevRunID uint           `json:"prevRunId"`
	NextRunID uint           `json:"nextRunId"`
	// State says whether the run still executes. The run model has no
	// status of its own: "running" and "finished" come from the run
	// tracker, and "not_running" is a run this master didn't start (e.g.
	// before a restart).
	State runs.RunState `json:"state"`
	// Devices are the IDs of the devices a running run holds now; empty
	// once the run is over.
	Devices []uint `json:"devices"`
}

// CheckRun tells that the test belongs to the app and the project, and the
// run to the test. Anything else is not found.
func (s *Service) CheckRun(ctx context.Context, ref testsvc.TestRef, runID uint) error {
	if err := s.deps.Repo.TestInScope(ctx, ref); err != nil {
		return err
	}
	return s.deps.Repo.RunInTest(ctx, ref.TestID, runID)
}

// Get loads a run of a test of the app with its startup history.
func (s *Service) Get(ctx context.Context, key Key) (*Page, error) {
	if err := s.CheckRun(ctx, key.TestRef, key.RunID); err != nil {
		return nil, err
	}
	loaded, err := s.deps.Repo.Run(ctx, key.RunID)
	if err != nil {
		return nil, err
	}
	if err := s.withStartupHistory(ctx, key.AppID, loaded.DeviceStatus); err != nil {
		return nil, err
	}
	return s.page(ctx, key.TestID, loaded)
}

// Last loads the newest run of a test of the app. A test without runs is an
// internal error, as it has always been answered.
func (s *Service) Last(ctx context.Context, ref testsvc.TestRef) (*Page, error) {
	if err := s.deps.Repo.TestInScope(ctx, ref); err != nil {
		return nil, err
	}
	loaded, err := s.deps.Repo.LastRun(ctx, ref.TestID)
	if apperr.KindOf(err) == apperr.KindNotFound {
		return nil, apperr.Internal("%v", err)
	}
	if err != nil {
		return nil, err
	}
	return s.page(ctx, ref.TestID, loaded)
}

// page adds the neighbours, the state and the held devices to a run.
func (s *Service) page(ctx context.Context, testID uint, loaded *models.TestRun) (*Page, error) {
	around, err := s.deps.Repo.Neighbours(ctx, testID, loaded.ID)
	if err != nil {
		return nil, err
	}
	for i := range loaded.Protocols {
		protocol.Clean(&loaded.Protocols[i])
	}
	return &Page{
		TestRun: *loaded, PrevRunID: around.Prev, NextRunID: around.Next,
		State: s.deps.RunState(loaded.ID), Devices: s.heldDevices(loaded.ID),
	}, nil
}

// heldDevices returns the devices an active run of this master holds.
func (s *Service) heldDevices(runID uint) []uint {
	runner, ok := s.deps.Runners.Get(runID)
	if !ok {
		return []uint{}
	}
	return runner.LockedDevices()
}

// withStartupHistory sets the average startup time of the earlier runs on
// each device status.
func (s *Service) withStartupHistory(ctx context.Context, appID uint, statuses []models.TestRunDeviceStatus) error {
	for i := range statuses {
		earlier, err := s.deps.Repo.StartupTimes(ctx, StartupQuery{AppID: appID, Status: &statuses[i], Limit: protocol.HistoryLimit})
		if err != nil {
			return err
		}
		var total uint
		for _, e := range earlier {
			total += e.StartupTime
		}
		if len(earlier) > 0 {
			statuses[i].HistAvgStartupTime = total / uint(len(earlier))
		}
	}
	return nil
}
