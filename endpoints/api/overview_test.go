package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// The overview only counts the project's own apps, bundles and tests, lists
// the tests whose latest run needs attention, what runs now, and the latest
// runs.
func TestOverview(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.AppBinary{}, &models.Test{}, &models.TestRun{},
		&models.TestProtocol{}, &models.TestRunLogEntry{}); err != nil {
		t.Fatal(err)
	}
	mine := models.Project{Identifier: "mine"}
	other := models.Project{Identifier: "other"}
	empty := models.Project{Identifier: "empty"}
	db.Create(&mine)
	db.Create(&other)
	db.Create(&empty)
	game := models.App{ProjectID: mine.ID, Name: "game"}
	theirs := models.App{ProjectID: other.ID, Name: "theirs"}
	db.Create(&game)
	db.Create(&models.App{ProjectID: mine.ID, Name: "editor"})
	db.Create(&theirs)
	db.Create(&models.AppBinary{AppID: game.ID, Size: 100})
	db.Create(&models.AppBinary{AppID: game.ID, Size: 50})
	db.Create(&models.AppBinary{AppID: theirs.ID, Size: 1000})

	test := func(app models.App, name string) models.Test {
		x := models.Test{AppID: app.ID, Name: name}
		db.Create(&x)
		return x
	}
	run := func(t models.Test, results ...models.TestResultState) uint {
		r := models.TestRun{TestID: t.ID}
		db.Create(&r)
		for _, res := range results {
			db.Create(&models.TestProtocol{TestRunID: r.ID, TestName: "S/A", TestResult: res})
		}
		return r.ID
	}
	smoke, payment, nightly, fresh := test(game, "smoke"), test(game, "payment"), test(game, "nightly"), test(game, "fresh")
	foreign := test(theirs, "foreign")

	run(smoke, models.TestResultSuccess)
	smokeFailed := run(smoke, models.TestResultFailed) // latest: failing
	paymentFailed := run(payment, models.TestResultFailed)
	run(payment, models.TestResultSuccess) // latest: passing again
	nightlyRunning := run(nightly, models.TestResultOpen)
	run(foreign, models.TestResultFailed)
	_ = fresh
	for i := 0; i < overviewLatestRuns; i++ {
		run(payment, models.TestResultSuccess)
	}

	gin.SetMode(gin.TestMode)
	s := &Service{db: db, runTracker: runs.NewTracker(nil)}
	s.runTracker.Started(nightlyRunning)
	r := gin.New()
	r.GET("/:project/overview", func(c *gin.Context) {
		p := mine
		if c.Param("project") == "empty" {
			p = empty
		}
		s.getOverview(c, &p)
	})
	get := func(path string) projectOverview {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		var o projectOverview
		if err := json.Unmarshal(w.Body.Bytes(), &o); err != nil {
			t.Fatal(err)
		}
		return o
	}

	o := get("/mine/overview")
	if o.Apps != 2 || o.Tests != 4 || o.Bundles != 2 || o.BundleStorage != 150 {
		t.Errorf("counts: apps %d tests %d bundles %d storage %d", o.Apps, o.Tests, o.Bundles, o.BundleStorage)
	}
	if len(o.Failing) != 1 || o.Failing[0].ID != smokeFailed || o.Failing[0].TestName != "smoke" || o.Failing[0].AppID != game.ID {
		t.Errorf("failing %+v, want only smoke's latest run", o.Failing)
	}
	if len(o.Running) != 1 || o.Running[0].ID != nightlyRunning || o.Running[0].Status != "running" {
		t.Errorf("running %+v", o.Running)
	}
	if len(o.Latest) != overviewLatestRuns || o.Latest[0].TestName != "payment" {
		t.Errorf("latest: %d runs, first %+v", len(o.Latest), o.Latest)
	}
	for _, l := range o.Latest {
		if l.ID == paymentFailed || l.TestName == "foreign" {
			t.Errorf("latest has an old or a foreign run: %+v", l)
		}
	}

	e := get("/empty/overview")
	if e.Apps != 0 || e.Tests != 0 || e.Failing == nil || len(e.Latest) != 0 {
		t.Errorf("empty project %+v", e)
	}
}
