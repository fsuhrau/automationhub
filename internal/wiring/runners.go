package wiring

import (
	"github.com/fsuhrau/automationhub/apperr"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
)

func newRunService(in Infra, rt *runtime, tests *testsvc.Service) *runsvc.Service {
	return runsvc.New(runsvc.Deps{
		Repo:       runstore.New(in.DB),
		Pool:       in.Devices,
		Runners:    rt.runners,
		NewRunner:  runnerFactory(in),
		Summarizer: tests,
		RunState:   rt.runState,
		Started:    rt.tracker.Started,
	})
}

// runnerFactory builds the runner of a test configuration type on the
// database, the device pool and the folders of the config.
func runnerFactory(in Infra) runsvc.NewRunner {
	return func(kind models.TestType, project *models.Project, appID uint) (runsvc.Runner, error) {
		deps := base.Deps{
			DB:        in.DB,
			Devices:   in.Devices,
			NodeURL:   in.Config.NodeUrl,
			DataDir:   in.Config.Paths.TestData,
			StoreDir:  in.Config.Paths.AppStore,
			ProjectID: project.Identifier,
			AppID:     appID,
		}
		return newRunner(kind, deps)
	}
}

func newRunner(kind models.TestType, deps base.Deps) (runsvc.Runner, error) {
	switch kind {
	case models.TestTypeUnity:
		return unity.NewWithDeps(deps), nil
	case models.TestTypeScenario:
		return scenario.NewWithDeps(deps), nil
	}
	return nil, apperr.Internal("invalid test config type")
}
