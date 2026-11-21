package teststore_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"gorm.io/gorm"
)

func bg() context.Context { return context.Background() }

// world is two projects with one app and test each; the first test has a
// full Unity config.
type world struct {
	db            *gorm.DB
	store         *teststore.Store
	mine, foreign *models.Test
	scope         service.Scope
	device        *models.Device
}

func newWorld(t *testing.T) *world {
	t.Helper()
	db := testkit.DB(t)
	w := &world{db: db, store: teststore.New(db)}
	foreignApp := testkit.App(db)
	w.foreign = testkit.Test(db, func(x *models.Test) { x.AppID = foreignApp.ID; x.Name = "foreign" })
	project := testkit.Project(db, func(p *models.Project) { p.CompanyID = 7 })
	app := testkit.App(db, func(a *models.App) { a.ProjectID = project.ID })
	w.mine = testkit.Test(db, func(x *models.Test) { x.AppID = app.ID; x.Name = "mine" })
	w.scope = service.Scope{ProjectID: project.ID, AppID: app.ID}
	w.device = testkit.Device(db, func(d *models.Device) { d.CompanyID = 7 })
	cfg := models.TestConfig{TestID: w.mine.ID, Type: models.TestTypeUnity}
	db.Create(&cfg)
	db.Create(&models.TestConfigDevice{TestConfigID: cfg.ID, DeviceID: w.device.ID})
	unity := models.TestConfigUnity{TestConfigID: cfg.ID}
	db.Create(&unity)
	db.Create(&models.UnityTestFunction{TestConfigUnityID: unity.ID, Method: "Old"})
	return w
}

func (w *world) ref(id uint) service.TestRef {
	return service.TestRef{Scope: w.scope, TestID: id}
}

func (w *world) drop(t *testing.T, table string) {
	t.Helper()
	if err := w.db.Migrator().DropTable(table); err != nil {
		t.Fatal(err)
	}
}

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != kind {
		t.Fatalf("error %v has kind %d, want %d", err, got, kind)
	}
}

func TestTestsListsTheScopeWithPreloadedConfig(t *testing.T) {
	w := newWorld(t)

	tests, err := w.store.Tests(bg(), w.scope)

	if err != nil || len(tests) != 1 || tests[0].ID != w.mine.ID {
		t.Fatalf("tests %+v, err %v; want only %d", tests, err, w.mine.ID)
	}
	cfg := tests[0].TestConfig
	if len(cfg.Devices) != 1 || cfg.Unity == nil || len(cfg.Unity.UnityTestFunctions) != 1 {
		t.Fatalf("config not preloaded: %+v", cfg)
	}
}

func TestTestsOfAnotherProjectsAppIsEmpty(t *testing.T) {
	w := newWorld(t)
	sc := service.Scope{ProjectID: w.scope.ProjectID + 99, AppID: w.scope.AppID}

	tests, err := w.store.Tests(bg(), sc)

	if err != nil || len(tests) != 0 {
		t.Fatalf("tests %+v, err %v; want none", tests, err)
	}
}

func TestTestsMapsDatabaseErrors(t *testing.T) {
	w := newWorld(t)
	w.drop(t, "tests")

	_, err := w.store.Tests(bg(), w.scope)

	wantKind(t, err, apperr.KindInternal)
}

func TestTestPreloadsTheDeviceBehindTheConfig(t *testing.T) {
	w := newWorld(t)

	test, err := w.store.Test(bg(), w.ref(w.mine.ID))

	if err != nil || test.TestConfig.Devices[0].Device.ID != w.device.ID {
		t.Fatalf("test %+v, err %v", test, err)
	}
}

func TestTestNotFound(t *testing.T) {
	tests := []struct {
		name string
		sc   func(*world) service.Scope
		id   func(*world) uint
	}{
		{"unknown id", func(w *world) service.Scope { return w.scope }, func(*world) uint { return 999 }},
		{"another project's test", func(w *world) service.Scope { return w.scope }, func(w *world) uint { return w.foreign.ID }},
		{"own test through another project", func(w *world) service.Scope {
			return service.Scope{ProjectID: w.scope.ProjectID + 99, AppID: w.scope.AppID}
		}, func(w *world) uint { return w.mine.ID }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)

			_, err := w.store.Test(bg(), service.TestRef{Scope: tc.sc(w), TestID: tc.id(w)})

			wantKind(t, err, apperr.KindNotFound)
			if err.Error() != "record not found" {
				t.Fatalf("message %q, want the api's record not found", err)
			}
		})
	}
}

func TestTestMapsDatabaseErrors(t *testing.T) {
	w := newWorld(t)
	w.drop(t, "tests")

	_, err := w.store.Test(bg(), w.ref(w.mine.ID))

	wantKind(t, err, apperr.KindInternal)
}

func TestAppExists(t *testing.T) {
	w := newWorld(t)
	tests := []struct {
		name string
		sc   service.Scope
		want bool
	}{
		{"own app", w.scope, true},
		{"app of another project", service.Scope{ProjectID: w.scope.ProjectID + 99, AppID: w.scope.AppID}, false},
		{"unknown app", service.Scope{ProjectID: w.scope.ProjectID, AppID: 999}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := w.store.AppExists(bg(), tc.sc)

			if err != nil || got != tc.want {
				t.Fatalf("exists %v, err %v; want %v", got, err, tc.want)
			}
		})
	}
}

func TestAppExistsMapsDatabaseErrors(t *testing.T) {
	w := newWorld(t)
	w.drop(t, "apps")

	_, err := w.store.AppExists(bg(), w.scope)

	wantKind(t, err, apperr.KindInternal)
}

func TestCountDevicesCountsTheProjectsVisibleOnes(t *testing.T) {
	w := newWorld(t)
	shared := testkit.Device(w.db)
	foreignDevice := testkit.Device(w.db, func(d *models.Device) { d.CompanyID = 99 })
	ids := []uint{w.device.ID, shared.ID, foreignDevice.ID, 999}

	n, err := w.store.CountDevices(bg(), w.scope.ProjectID, ids)

	if err != nil || n != 2 {
		t.Fatalf("visible %d, err %v; want the company's and the shared device", n, err)
	}
}

func TestCountDevicesMapsDatabaseErrors(t *testing.T) {
	w := newWorld(t)
	w.drop(t, "devices")

	_, err := w.store.CountDevices(bg(), w.scope.ProjectID, []uint{1})

	wantKind(t, err, apperr.KindInternal)
}
