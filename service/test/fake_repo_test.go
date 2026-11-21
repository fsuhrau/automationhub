package test_test

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func bg() context.Context { return context.Background() }

// fakeRepo is an in-memory service.Repo. A test of the world belongs to the
// scope in tests; everything else is not found.
type fakeRepo struct {
	calls      []string
	failWith   map[string]error
	scope      service.Scope
	tests      []models.Test
	runs       []models.TestRun
	results    []service.ResultCount
	logErrors  []service.LogErrorCount
	visible    int
	appMissing bool
	created    models.Test
	updated    models.Test
	updatedOf  *models.Test
}

func (f *fakeRepo) step(name string) error {
	f.calls = append(f.calls, name)
	return f.failWith[name]
}

func (f *fakeRepo) Tests(_ context.Context, sc service.Scope) ([]models.Test, error) {
	if err := f.step("Tests"); err != nil {
		return nil, err
	}
	if sc != f.scope {
		return nil, nil
	}
	return f.tests, nil
}

func (f *fakeRepo) Test(_ context.Context, ref service.TestRef) (*models.Test, error) {
	if err := f.step("Test"); err != nil {
		return nil, err
	}
	for i := range f.tests {
		if f.tests[i].ID == ref.TestID && ref.Scope == f.scope {
			return &f.tests[i], nil
		}
	}
	return nil, apperr.NotFound("record not found")
}

func (f *fakeRepo) AppExists(_ context.Context, sc service.Scope) (bool, error) {
	if err := f.step("AppExists"); err != nil {
		return false, err
	}
	return sc == f.scope && !f.appMissing, nil
}

func (f *fakeRepo) CountDevices(context.Context, uint, []uint) (int, error) {
	return f.visible, f.step("CountDevices")
}

func (f *fakeRepo) CreateTest(_ context.Context, test models.Test) (models.Test, error) {
	f.created = test
	return test, f.step("CreateTest")
}

func (f *fakeRepo) UpdateTest(_ context.Context, current *models.Test, desired models.Test) error {
	f.updatedOf, f.updated = current, desired
	return f.step("UpdateTest")
}

func (f *fakeRepo) RecentRuns(context.Context, uint, int) ([]models.TestRun, error) {
	return f.runs, f.step("RecentRuns")
}

func (f *fakeRepo) RunResults(context.Context, []uint) ([]service.ResultCount, error) {
	return f.results, f.step("RunResults")
}

func (f *fakeRepo) RunLogErrors(context.Context, []uint) ([]service.LogErrorCount, error) {
	return f.logErrors, f.step("RunLogErrors")
}
