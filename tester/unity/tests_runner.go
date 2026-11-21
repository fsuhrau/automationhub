// Package unity runs the tests of a unity app: the app lists its tests, each
// is started on the devices and reports back its protocol.
package unity

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/base"
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

// TestList returns the tests the app lists or the config names.
func (s *strategy) TestList(ctx context.Context, connected []base.DeviceMap) ([]models.UnityTestFunction, error) {
	return s.runner.ListTests(ctx, connected)
}
