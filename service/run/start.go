package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
)

// Start starts the test with the bundle (or the start URL) on the devices
// of the test configuration. It fails right away when none of the devices
// can run tests, so the caller (web, CLI, MCP, CI) gets no empty run.
func (s *Service) Start(ctx context.Context, req StartRequest) (RunRef, error) {
	test, err := s.deps.Repo.StartTest(ctx, req.Project.ID, req.TestID)
	if err != nil {
		return RunRef{}, err
	}
	binary, err := s.bundleFor(ctx, test, req.BundleID)
	if err != nil {
		return RunRef{}, err
	}
	devices, err := s.devicesFor(ctx, req.Project, test)
	if err != nil {
		return RunRef{}, err
	}
	runner, err := s.runnerFor(test, req)
	if err != nil {
		return RunRef{}, err
	}
	started, err := runner.Run(devices, binary, req.StartURL)
	if err != nil {
		return RunRef{}, err
	}
	s.track(started, runner)
	announce(req.Project, test, started.ID)
	return RunRef{Run: started}, nil
}

// bundleFor loads the bundle of the app the test runs with; web and editor
// apps have none.
func (s *Service) bundleFor(ctx context.Context, test *models.Test, bundleID uint) (*models.AppBinary, error) {
	if platform := test.App.Platform; platform == models.PlatformTypeEditor || platform == models.PlatformTypeWeb {
		return nil, nil
	}
	binary, err := s.deps.Repo.Bundle(ctx, test.AppID, bundleID)
	if err != nil {
		return nil, err
	}
	if binary.Archived() {
		return nil, apperr.Invalid("app bundle %d is archived; upload its file again to run tests with it", bundleID)
	}
	return binary, nil
}

// devicesFor lists the devices the test runs on, each with its live device.
func (s *Service) devicesFor(ctx context.Context, project *models.Project, test *models.Test) ([]models.Device, error) {
	query := DeviceQuery{Project: project, All: test.TestConfig.AllDevices, IDs: test.TestConfig.GetDeviceIds()}
	devices, err := s.deps.Repo.Devices(ctx, query)
	if err != nil {
		return nil, err
	}
	for i := range devices {
		devices[i].Dev, _ = s.deps.Pool.GetDevice(devices[i].DeviceIdentifier)
	}
	return devices, checkDevices(devices)
}

// runnerFor builds the runner of the test's type and prepares it.
func (s *Service) runnerFor(test *models.Test, req StartRequest) (Runner, error) {
	kind := test.TestConfig.Type
	if kind == models.TestTypeUnity && test.TestConfig.Unity == nil {
		return nil, apperr.Internal("test config %d has no unity settings", test.TestConfig.ID)
	}
	runner, err := s.deps.NewRunner(kind, req.Project, test.AppID)
	if err != nil {
		return nil, err
	}
	return runner, runner.Initialize(*test, protocol.ExtractParams(req.Params))
}

// track remembers the runner of the run and tells the tracker that it
// started.
func (s *Service) track(started *models.TestRun, runner Runner) {
	s.deps.Runners.Add(started.ID, runner)
	s.deps.Started(started.ID)
}

// announce tells that a run of the test started.
func announce(project *models.Project, test *models.Test, runID uint) {
	events.TestRunStarted.Trigger(events.TestRunStartedPayload{
		ProjectID: project.Identifier,
		AppID:     test.AppID,
		TestID:    test.ID,
		TestRunID: runID,
	})
}

// checkDevices refuses a run when none of its devices can run tests.
// Unavailable devices next to available ones are skipped by the run.
func checkDevices(devices []models.Device) error {
	if len(devices) == 0 {
		return apperr.Conflict("no device available to run the test: the test has no devices")
	}
	reasons := make([]string, 0, len(devices))
	for i := range devices {
		reason := base.Unavailable(devices[i])
		if reason == "" {
			return nil
		}
		reasons = append(reasons, fmt.Sprintf("%s is %s", device.DisplayName(&devices[i]), reason))
	}
	return apperr.Conflict("no device available to run the test: %s", strings.Join(reasons, "; "))
}
