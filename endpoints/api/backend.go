package api

import (
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
	runsapi "github.com/fsuhrau/automationhub/endpoints/api/runs"
	"github.com/fsuhrau/automationhub/hub/runs"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
)

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
