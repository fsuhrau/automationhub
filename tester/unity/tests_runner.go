// Package unity runs the tests of a unity app: the app lists its tests, each
// is started on the devices and reports back its protocol.
package unity

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/base"
	"gorm.io/gorm"
)

const (
	DefaultTestTimeout = 5 * time.Minute
)

// strategy is the unity part of the pipeline.
type strategy struct {
	runner *base.Runner
}

var (
	_ base.Strategy = (*strategy)(nil)
	_ base.Preparer = (*strategy)(nil)
)

// NewWithDeps builds the runner of unity tests.
func NewWithDeps(deps base.Deps) tester.Interface {
	return base.NewRunner(deps, models.TestTypeUnity, func(r *base.Runner) base.Strategy {
		return &strategy{runner: r}
	})
}

// New builds the runner of unity tests.
//
// Deprecated: use NewWithDeps (G34 switches callers)
//
//nolint:revive // the signature is fixed by its callers until they use NewWithDeps
func New(db *gorm.DB, nodeUrl string, deviceManager manager.Devices, _ sse.Publisher, projectId string, appId uint) tester.Interface {
	return NewWithDeps(base.Deps{DB: db, NodeURL: nodeUrl, Devices: deviceManager, ProjectID: projectId, AppID: appId})
}

// TestList returns the tests the app lists or the config names.
func (s *strategy) TestList(ctx context.Context, connected []base.DeviceMap) ([]models.UnityTestFunction, error) {
	return s.runner.ListTests(ctx, connected)
}
