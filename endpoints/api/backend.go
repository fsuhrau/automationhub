package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
	runsapi "github.com/fsuhrau/automationhub/endpoints/api/runs"
	"github.com/fsuhrau/automationhub/hub/runs"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
)

// ErrNotFound marks lookups that failed because the entity doesn't exist or
// isn't part of the project.
var ErrNotFound = errors.New("not found")

// ErrBundleExists is returned when an identical app bundle was uploaded before.
var ErrBundleExists = errors.New("app bundle exists already")

// RunTestRequest is the body of a run request, for clients such as the CLI.
type RunTestRequest = runsapi.StartRequest

// runBook is the registry of the active runners of this master, kept in the
// fields of the Service.
type runBook struct{ s *Service }

func runKey(runID uint) string {
	return strconv.FormatUint(uint64(runID), 10)
}

func (b runBook) Add(runID uint, r runsvc.Runner) {
	b.s.runnersMutex.Lock()
	defer b.s.runnersMutex.Unlock()
	b.s.runners[runKey(runID)] = r
}

func (b runBook) Get(runID uint) (runsvc.Runner, bool) {
	b.s.runnersMutex.Lock()
	defer b.s.runnersMutex.Unlock()
	r, ok := b.s.runners[runKey(runID)]
	return r, ok
}

func (b runBook) Remove(runID uint) {
	b.s.runnersMutex.Lock()
	defer b.s.runnersMutex.Unlock()
	delete(b.s.runners, runKey(runID))
}

func (s *Service) onRunFinished(runID uint) {
	runBook{s}.Remove(runID)
}

// runState tells whether this master is running the run right now.
func (s *Service) runState(runID uint) runs.RunState {
	if s.runTracker == nil {
		return runs.RunUnknown
	}
	state, _ := s.runTracker.Status(runID)
	return state
}

// newRunner builds the runner of a test configuration type.
func (s *Service) newRunner(kind models.TestType, project *models.Project, appID uint) (runsvc.Runner, error) {
	switch kind {
	case models.TestTypeUnity:
		return unity.New(s.db, s.nodeUrl, s.devicesManager, s, project.Identifier, appID), nil //nolint:staticcheck // the brief keeps the old constructors until G34 switches callers
	case models.TestTypeScenario:
		return scenario.New(s.db, s.nodeUrl, s.devicesManager, s, project.Identifier, appID), nil //nolint:staticcheck // the brief keeps the old constructors until G34 switches callers
	}
	return nil, apperr.Internal("invalid test config type")
}

// runService builds the run use cases on the database, the device pool and
// the runners and run tracker of this master.
func (s *Service) runService() *runsvc.Service {
	return runsvc.New(runsvc.Deps{
		Repo:       runstore.New(s.db),
		Pool:       s.devicesManager,
		Runners:    runBook{s},
		NewRunner:  s.newRunner,
		Summarizer: s.testService(),
		RunState:   s.runState,
		Started:    s.runTracker.Started,
	})
}

// ProjectDevices lists the devices visible to the project.
func (s *Service) ProjectDevices(project *models.Project) ([]models.Device, error) {
	return s.runService().Devices(context.Background(), project)
}

// FindDevice returns one device of the project.
func (s *Service) FindDevice(project *models.Project, id uint) (*models.Device, error) {
	d, err := s.runService().Device(context.Background(), project, id)
	if apperr.KindOf(err) == apperr.KindNotFound {
		return nil, fmt.Errorf("%w: device %d", ErrNotFound, id)
	}
	return d, err
}

// StartTestRun starts the test with the given bundle (or start URL for web
// and editor apps) on the devices of the test configuration.
func (s *Service) StartTestRun(project *models.Project, testID, binaryID uint, startURL, params string) (*models.TestRun, error) {
	ref, err := s.runService().Start(context.Background(), runsvc.StartRequest{
		Project: project, TestID: testID, BundleID: binaryID, StartURL: startURL, Params: params,
	})
	return ref.Run, err
}

// CancelTestRun cancels a run started by this master.
func (s *Service) CancelTestRun(project *models.Project, runID uint) error {
	return s.runService().Cancel(context.Background(), project.ID, runID)
}

// StoreBundle analyses an uploaded app file, moves it into the app storage and
// creates the bundle entry. filePath must be inside apps.AppStoragePath. An
// identical bundle comes back with ErrBundleExists.
func (s *Service) StoreBundle(application *models.App, filePath string) (*models.AppBinary, error) {
	scope := appsvc.Scope{ProjectID: application.ProjectID, AppID: application.ID}
	bundle, err := s.appService().StoreBundle(context.Background(), scope, filePath)
	if errors.Is(err, appsvc.ErrBundleExists) {
		return bundle, ErrBundleExists
	}
	return bundle, err
}
