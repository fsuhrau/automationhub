package unity

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
)

// strategyFixture is a unity strategy of a started run on one device.
type strategyFixture struct {
	*basetest.Setup
	runner   *base.Runner
	strategy *strategy
}

func newStrategyFixture(t *testing.T) *strategyFixture {
	t.Helper()
	f := &strategyFixture{Setup: basetest.NewSetup(t, 1)}
	f.runner = base.NewRunner(f.Deps, models.TestTypeUnity, func(r *base.Runner) base.Strategy {
		f.strategy = &strategy{runner: r}
		return f.strategy
	})
	f.Test.TestConfig = models.TestConfig{Type: models.TestTypeUnity}
	if err := f.runner.Initialize(*f.Test, map[string]string{"KEY": "value"}); err != nil {
		t.Fatal(err)
	}
	if err := f.runner.InitNewTestSession(context.Background(), base.Session{}); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *strategyFixture) logged(level, text string) bool {
	var n int64
	f.DB.Model(&models.TestRunLogEntry{}).
		Where("test_run_id = ? AND level = ? AND log LIKE ?", f.runner.TestRun.ID, level, "%"+text+"%").Count(&n)
	return n > 0
}

func (f *strategyFixture) protocols() []models.TestProtocol {
	var out []models.TestProtocol
	f.DB.Where("test_run_id = ?", f.runner.TestRun.ID).Find(&out)
	return out
}
