package test_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func validEdit() service.Edit {
	return service.Edit{
		Name: " Nightly ", ExecutionType: models.SimultaneouslyExecutionType, RecordSession: true,
		Devices: []uint{4, 3, 4}, UnityTestCategoryType: models.AllOfCategory, Categories: " a, ,a ",
		TestFunctions: []models.UnityTestFunction{{Class: "S", Method: "New"}, {Class: "S", Method: "New"}, {Method: " "}},
	}
}

func TestCreateStoresTheTidiedTest(t *testing.T) {
	repo := newFake()
	repo.visible = 2
	svc := service.New(service.Deps{Repo: repo})

	created, err := svc.Create(bg(), mine(), service.NewTest{Type: models.TestTypeUnity, Edit: validEdit()})

	if err != nil || created.Name != "Nightly" || created.AppID != 2 {
		t.Fatalf("created %+v, err %v", created, err)
	}
	cfg := repo.created.TestConfig
	if !reflect.DeepEqual(cfg.Devices, []models.TestConfigDevice{{DeviceID: 4}, {DeviceID: 3}}) ||
		cfg.Unity.Categories != "a" || len(cfg.Unity.UnityTestFunctions) != 1 || !cfg.RecordSession || cfg.Type != models.TestTypeUnity {
		t.Fatalf("stored config %+v", cfg)
	}
}

func TestCreateRefusals(t *testing.T) {
	tests := []struct {
		name  string
		sc    service.Scope
		input func(*service.NewTest)
		setup func(*fakeRepo)
		kind  apperr.Kind
	}{
		{"scenario test", mine(), func(n *service.NewTest) { n.Type = models.TestTypeScenario }, nil, apperr.KindInvalid},
		{"invalid edit", mine(), func(n *service.NewTest) { n.Name = " " }, nil, apperr.KindInvalid},
		{"app of another project", foreign(), nil, nil, apperr.KindNotFound},
		{"unknown devices", mine(), nil, func(f *fakeRepo) { f.visible = 1 }, apperr.KindInvalid},
		{"app lookup fails", mine(), nil, func(f *fakeRepo) { f.failWith["AppExists"] = boom }, apperr.KindInternal},
		{"device count fails", mine(), nil, func(f *fakeRepo) { f.failWith["CountDevices"] = boom; f.visible = 2 }, apperr.KindInternal},
		{"store fails", mine(), nil, func(f *fakeRepo) { f.visible = 2; f.failWith["CreateTest"] = boom }, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFake()
			if tc.setup != nil {
				tc.setup(repo)
			}
			in := service.NewTest{Type: models.TestTypeUnity, Edit: validEdit()}
			if tc.input != nil {
				tc.input(&in)
			}

			_, err := service.New(service.Deps{Repo: repo}).Create(bg(), tc.sc, in)

			wantKind(t, err, tc.kind)
		})
	}
}

func TestUpdateSavesDesiredConfigOverTheLoadedTest(t *testing.T) {
	repo := newFake()
	repo.visible = 2
	svc := service.New(service.Deps{Repo: repo})

	_, err := svc.Update(bg(), ref(mine(), 1), validEdit())

	if err != nil || repo.updatedOf.ID != 1 || repo.updated.Name != "Nightly" {
		t.Fatalf("updated %+v of %+v, err %v", repo.updated, repo.updatedOf, err)
	}
	if repo.updated.TestConfig.Type != models.TestTypeUnity || repo.updated.TestConfig.Unity.Categories != "a" {
		t.Fatalf("desired config %+v", repo.updated.TestConfig)
	}
}

func TestUpdateOfATestWithoutUnitySettingsSkipsThem(t *testing.T) {
	repo := newFake()
	repo.visible = 2
	repo.tests[0].TestConfig.Unity = nil
	edit := validEdit()
	edit.Categories, edit.TestFunctions = "", nil

	_, err := service.New(service.Deps{Repo: repo}).Update(bg(), ref(mine(), 1), edit)

	if err != nil || repo.updated.TestConfig.Unity != nil {
		t.Fatalf("desired config %+v, err %v", repo.updated.TestConfig, err)
	}
}

func TestUpdateRefusals(t *testing.T) {
	tests := []struct {
		name  string
		sc    service.Scope
		id    uint
		input func(*service.Edit)
		setup func(*fakeRepo)
		kind  apperr.Kind
	}{
		{"test of another project", foreign(), 1, nil, nil, apperr.KindNotFound},
		{"unknown test", mine(), 8, nil, nil, apperr.KindNotFound},
		{"invalid edit", mine(), 1, func(e *service.Edit) { e.Name = "" }, nil, apperr.KindInvalid},
		{"unknown devices", mine(), 1, nil, func(f *fakeRepo) { f.visible = 1 }, apperr.KindInvalid},
		{"device count fails", mine(), 1, nil, func(f *fakeRepo) { f.failWith["CountDevices"] = boom }, apperr.KindInternal},
		{"transaction fails", mine(), 1, nil, func(f *fakeRepo) { f.visible = 2; f.failWith["UpdateTest"] = boom }, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFake()
			repo.visible = 2
			if tc.setup != nil {
				tc.setup(repo)
			}
			edit := validEdit()
			if tc.input != nil {
				tc.input(&edit)
			}

			_, err := service.New(service.Deps{Repo: repo}).Update(bg(), ref(tc.sc, tc.id), edit)

			wantKind(t, err, tc.kind)
		})
	}
}

func TestUpdateReportsAFailingReload(t *testing.T) {
	repo := newFake()
	repo.visible = 2
	repo.failWith["UpdateTest"] = nil
	svc := service.New(service.Deps{Repo: &reloadFails{fakeRepo: repo}})

	_, err := svc.Update(bg(), ref(mine(), 1), validEdit())

	wantKind(t, err, apperr.KindInternal)
}

// reloadFails loads the test once and fails the reload after the update.
type reloadFails struct {
	*fakeRepo
	loads int
}

func (r *reloadFails) Test(c context.Context, ref service.TestRef) (*models.Test, error) {
	r.loads++
	if r.loads > 1 {
		return nil, boom
	}
	return r.fakeRepo.Test(c, ref)
}
