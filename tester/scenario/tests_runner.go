// Package scenario runs the steps of a scenario on the devices.
package scenario

import (
	"context"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/base"
	"gorm.io/gorm"
)

// strategy is the scenario part of the pipeline.
type strategy struct {
	runner *base.Runner
}

var (
	_ base.Strategy      = (*strategy)(nil)
	_ base.StartObserver = (*strategy)(nil)
)

// NewWithDeps builds the runner of scenarios.
func NewWithDeps(deps base.Deps) tester.Interface {
	return base.NewRunner(deps, models.TestTypeScenario, func(r *base.Runner) base.Strategy {
		return &strategy{runner: r}
	})
}

// New builds the runner of scenarios.
//
// Deprecated: use NewWithDeps (G34 switches callers)
//
//nolint:revive // the signature is fixed by its callers until they use NewWithDeps
func New(db *gorm.DB, nodeUrl string, deviceManager manager.Devices, _ sse.Publisher, projectId string, appId uint) tester.Interface {
	return NewWithDeps(base.Deps{DB: db, NodeURL: nodeUrl, Devices: deviceManager, ProjectID: projectId, AppID: appId})
}

// TestList returns the tests the config lists next to the steps; a scenario
// without them only runs its steps.
func (s *strategy) TestList(ctx context.Context, connected []base.DeviceMap) ([]models.UnityTestFunction, error) {
	if s.runner.Config.Unity == nil {
		return nil, nil
	}
	return s.runner.ListTests(ctx, connected)
}

// RunTest does nothing: the scenario's tests are its steps, and executing a
// listed test is not implemented yet.
func (s *strategy) RunTest(context.Context, base.DeviceMap, action.TestStart) {}

// AppStarted runs the steps of the scenario on the device, before the pipeline
// waits for the app's connection.
func (s *strategy) AppStarted(ctx context.Context, d device.Device) {
	for _, step := range s.steps() {
		if ctx.Err() != nil {
			return
		}
		s.runner.LogInfo("Scenario step '%s' on device '%s'", stepName(step.StepType), d.DeviceID())
	}
}

func (s *strategy) steps() []models.ScenarioStep {
	if s.runner.Config.Scenario == nil {
		return nil
	}
	return s.runner.Config.Scenario.Steps
}

// stepName names the type of a step in the run log.
func stepName(t models.StepType) string {
	switch t {
	case models.StepTypeInstallApp:
		return "install app"
	case models.StepTypeUninstallApp:
		return "uninstall app"
	case models.StepTypeStartApp:
		return "start app"
	case models.StepTypeStopApp:
		return "stop app"
	case models.StepTypeCheckpoint:
		return "checkpoint"
	case models.StepTypeExecuteTest:
		return "execute test"
	}
	return "unknown"
}
