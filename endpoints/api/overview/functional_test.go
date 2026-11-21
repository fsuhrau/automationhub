package overview_test

import (
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/runfixture"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/internal/testkit"
	svc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/models"
)

// overviewWorld is the project "mine" with the tests smoke, payment and
// nightly, and the runs that make smoke fail and nightly open.
type overviewWorld struct {
	*runfixture.World
	smokeFailed, paymentFailed uint
	empty                      *models.Project
}

func newOverviewWorld(t *testing.T) *overviewWorld {
	t.Helper()
	w := &overviewWorld{World: runfixture.New(t)}
	db := w.DB
	w.empty = testkit.Project(db, func(p *models.Project) { p.Identifier = "empty" })
	testkit.App(db, func(a *models.App) { a.ProjectID = w.MineApp.ProjectID })
	db.Create(&models.AppBinary{AppID: w.MineApp.ID, Size: 100})
	db.Create(&models.AppBinary{AppID: w.MineApp.ID, Size: 50})
	db.Create(&models.AppBinary{AppID: w.ForeignApp.ID, Size: 1000})
	payment := testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.MineApp.ID, "payment" })
	nightly := testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = w.MineApp.ID, "nightly" })
	w.smokeFailed = w.run(w.Mine, models.TestResultFailed) // the latest run of smoke: failing
	w.paymentFailed = w.run(payment, models.TestResultFailed)
	w.run(payment, models.TestResultSuccess) // latest: passing again
	w.run(nightly, models.TestResultOpen)
	w.run(w.Foreign, models.TestResultFailed)
	for i := 0; i < svc.LatestRuns; i++ {
		w.run(payment, models.TestResultSuccess)
	}
	return w
}

func (w *overviewWorld) run(test *models.Test, results ...models.TestResultState) uint {
	r := testkit.TestRun(w.DB, func(r *models.TestRun) { r.TestID = test.ID })
	for _, res := range results {
		testkit.Protocol(w.DB, func(p *models.TestProtocol) { p.TestRunID, p.TestResult = r.ID, res })
	}
	return r.ID
}

// The overview only counts the project's own apps, bundles and tests, lists
// the tests whose latest run needs attention, and the latest runs.
func TestOverviewCountsOnlyTheProjectAndListsWhatNeedsAttention(t *testing.T) {
	w := newOverviewWorld(t)

	o := testkit.Decode[svc.Overview](t, w.Do("GET", "/api/mine/overview", nil))

	if o.Apps != 2 || o.Tests != 3 || o.Bundles != 2 || o.BundleStorage != 150 {
		t.Errorf("counts: apps %d tests %d bundles %d storage %d", o.Apps, o.Tests, o.Bundles, o.BundleStorage)
	}
	if len(o.Failing) != 1 || o.Failing[0].ID != w.smokeFailed || o.Failing[0].TestName != "smoke" || o.Failing[0].AppID != w.MineApp.ID {
		t.Errorf("failing %+v, want only smoke's latest run", o.Failing)
	}
	if len(o.Running) != 0 {
		t.Errorf("running %+v, nothing runs on this master", o.Running)
	}
}

func TestOverviewLatestRunsAreTheNewestOfTheProject(t *testing.T) {
	w := newOverviewWorld(t)

	o := testkit.Decode[svc.Overview](t, w.Do("GET", "/api/mine/overview", nil))

	if len(o.Latest) != svc.LatestRuns || o.Latest[0].TestName != "payment" {
		t.Fatalf("latest: %d runs, first %+v", len(o.Latest), o.Latest)
	}
	for _, l := range o.Latest {
		if l.ID == w.paymentFailed || l.TestName == "foreign" || l.State != runs.RunUnknown {
			t.Errorf("latest has an old or a foreign run: %+v", l)
		}
	}
}

func TestOverviewOfAProjectWithoutAppsIsEmpty(t *testing.T) {
	w := newOverviewWorld(t)

	e := testkit.Decode[svc.Overview](t, w.Do("GET", "/api/"+w.empty.Identifier+"/overview", nil))

	if e.Apps != 0 || e.Tests != 0 || e.Failing == nil || len(e.Latest) != 0 {
		t.Errorf("empty project %+v", e)
	}
}

// Regression: /stats counted the runs, bundles and devices of all projects.
func TestStatsAreScopedToTheProject(t *testing.T) {
	w := runfixture.New(t)
	db := w.DB
	db.Create(&models.AppBinary{AppID: w.MineApp.ID, Size: 100})
	db.Create(&models.AppBinary{AppID: w.ForeignApp.ID, Size: 1000})
	testkit.Device(db, func(d *models.Device) { d.CompanyID = 2 })
	db.Model(w.ForeProto).Update("test_result", models.TestResultFailed)

	mine := testkit.Decode[svc.Stats](t, w.Do("GET", "/api/mine/stats", nil))
	foreign := testkit.Decode[svc.Stats](t, w.Do("GET", "/api/foreign/stats", nil))

	if mine.AppsCount != 1 || mine.AppsStorageSize != 100 || mine.DeviceCount != 1 {
		t.Errorf("mine counts %+v; the foreign bundle and the other company's device must not count", mine)
	}
	if len(mine.TestsLastProtocols) != 2 || len(mine.TestsLastFailed) != 0 {
		t.Errorf("mine protocols %d, failed %d; the foreign run must not count", len(mine.TestsLastProtocols), len(mine.TestsLastFailed))
	}
	if foreign.AppsCount != 1 || foreign.AppsStorageSize != 1000 || len(foreign.TestsLastProtocols) != 1 || len(foreign.TestsLastFailed) != 1 {
		t.Errorf("foreign stats %+v", foreign)
	}
}

func TestStatsOfAnUnknownProjectIsNotFound(t *testing.T) {
	w := runfixture.New(t)

	if rec := w.Do("GET", "/api/nobody/stats", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}
