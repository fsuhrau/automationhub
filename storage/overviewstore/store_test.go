package overviewstore_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/overviewstore"
	"gorm.io/gorm"
)

func bg() context.Context { return context.Background() }

// world is the project "mine" with two apps, two bundles and two tests, and
// a foreign project with an app, a bundle, a test and a failed run.
type world struct {
	db            *gorm.DB
	store         *overviewstore.Store
	mine, foreign *models.Project
	game, editor  *models.App
	smoke, nightly,
	theirs *models.Test
}

func newWorld(t *testing.T) *world {
	t.Helper()
	db := testkit.DB(t)
	w := &world{db: db, store: overviewstore.New(db)}
	w.mine = testkit.Project(db)
	w.foreign = testkit.Project(db)
	w.game = testkit.App(db, func(a *models.App) { a.ProjectID = w.mine.ID })
	w.editor = testkit.App(db, func(a *models.App) { a.ProjectID = w.mine.ID })
	foreignApp := testkit.App(db, func(a *models.App) { a.ProjectID = w.foreign.ID })
	db.Create(&models.AppBinary{AppID: w.game.ID, Size: 100})
	archived := models.AppBinary{AppID: w.game.ID, Size: 50}
	db.Create(&archived)
	db.Model(&archived).Update("archived_at", gorm.Expr("CURRENT_TIMESTAMP"))
	db.Create(&models.AppBinary{AppID: foreignApp.ID, Size: 1000})
	w.smoke = testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.game.ID, "smoke" })
	w.nightly = testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.editor.ID, "nightly" })
	w.theirs = testkit.Test(db, func(x *models.Test) { x.AppID = foreignApp.ID })
	return w
}

func (w *world) run(test *models.Test, results ...models.TestResultState) *models.TestRun {
	r := testkit.TestRun(w.db, func(r *models.TestRun) { r.TestID = test.ID })
	for _, res := range results {
		testkit.Protocol(w.db, func(p *models.TestProtocol) { p.TestRunID, p.TestResult = r.ID, res })
	}
	return r
}

func TestTotalsCountOnlyTheProjectsAppsAndBundlesAndNoArchivedStorage(t *testing.T) {
	w := newWorld(t)

	got, err := w.store.Totals(bg(), w.mine.ID)

	want := overview.Totals{Apps: 2, Bundles: 2, BundleStorage: 100}
	if err != nil || got != want {
		t.Fatalf("totals %+v, err %v; want %+v", got, err, want)
	}
}

func TestTestsAreTheOnesOfTheProjectsAppsById(t *testing.T) {
	w := newWorld(t)

	got, err := w.store.Tests(bg(), w.mine.ID)

	if err != nil || len(got) != 2 || got[0].ID != w.smoke.ID || got[1].ID != w.nightly.ID {
		t.Fatalf("tests %+v, err %v", got, err)
	}
}

func TestLastRunsHaveTheNewestRunOfEachTestNewestFirst(t *testing.T) {
	w := newWorld(t)
	w.run(w.smoke)
	smokeLast := w.run(w.smoke)
	nightlyLast := w.run(w.nightly)
	w.run(w.theirs)

	got, err := w.store.LastRuns(bg(), []uint{w.smoke.ID, w.nightly.ID})

	if err != nil || len(got) != 2 || got[0].ID != nightlyLast.ID || got[1].ID != smokeLast.ID {
		t.Fatalf("last runs %+v, err %v", got, err)
	}
}

func TestLatestRunsAreLimitedAndOnlyOfTheTests(t *testing.T) {
	w := newWorld(t)
	first := w.run(w.smoke)
	second := w.run(w.smoke)
	w.run(w.theirs)

	got, err := w.store.LatestRuns(bg(), []uint{w.smoke.ID}, 1)

	if err != nil || len(got) != 1 || got[0].ID != second.ID {
		t.Fatalf("latest %+v, err %v; want %d and not %d", got, err, second.ID, first.ID)
	}
}

func TestProtocolsAreTheProjectsNewestWithTheirRunTestAndDevice(t *testing.T) {
	w := newWorld(t)
	device := testkit.Device(w.db)
	failedRun := w.run(w.smoke, models.TestResultFailed, models.TestResultSuccess)
	w.db.Model(&models.TestProtocol{}).Where("test_run_id = ?", failedRun.ID).Update("device_id", device.ID)
	w.run(w.theirs, models.TestResultFailed)

	all, err := w.store.Protocols(bg(), overview.ProtocolQuery{ProjectID: w.mine.ID, Limit: 10})
	if err != nil || len(all) != 2 {
		t.Fatalf("protocols %+v, err %v; the foreign project's must not count", all, err)
	}
	if all[0].TestRun == nil || all[0].TestRun.Test == nil || all[0].Device == nil {
		t.Errorf("protocol without its run, test or device: %+v", all[0])
	}
	failed, err := w.store.Protocols(bg(), overview.ProtocolQuery{ProjectID: w.mine.ID, FailedOnly: true, Limit: 10})
	if err != nil || len(failed) != 1 || failed[0].TestResult != models.TestResultFailed {
		t.Fatalf("failed %+v, err %v", failed, err)
	}
	limited, err := w.store.Protocols(bg(), overview.ProtocolQuery{ProjectID: w.mine.ID, Limit: 1})
	if err != nil || len(limited) != 1 {
		t.Fatalf("limited %+v, err %v", limited, err)
	}
}

func TestDevicesAreTheOnesTheProjectCanSee(t *testing.T) {
	w := newWorld(t)
	shared := testkit.Device(w.db)
	testkit.Device(w.db, func(d *models.Device) { d.CompanyID = 99 })

	got, err := w.store.Devices(bg(), w.mine)

	if err != nil || len(got) != 1 || got[0].ID != shared.ID {
		t.Fatalf("devices %+v, err %v", got, err)
	}
}

func TestReadsReportADatabaseError(t *testing.T) {
	db := testkit.DB(t)
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	s := overviewstore.New(db)
	project := &models.Project{}
	_, e1 := s.Totals(bg(), 1)
	_, e2 := s.Tests(bg(), 1)
	_, e3 := s.LastRuns(bg(), []uint{1})
	_, e4 := s.LatestRuns(bg(), []uint{1}, 1)
	_, e5 := s.Protocols(bg(), overview.ProtocolQuery{ProjectID: 1, Limit: 1})
	_, e6 := s.Devices(bg(), project)
	for i, err := range []error{e1, e2, e3, e4, e5, e6} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("read %d: %v", i, err)
		}
	}
}

func TestTotalsReportAFailingBundleQuery(t *testing.T) {
	w := newWorld(t)
	if err := w.db.Migrator().DropTable(&models.AppBinary{}); err != nil {
		t.Fatal(err)
	}

	_, err := w.store.Totals(bg(), w.mine.ID)

	if apperr.KindOf(err) != apperr.KindInternal {
		t.Fatalf("error %v", err)
	}
}
