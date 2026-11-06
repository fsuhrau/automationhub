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

func TestSummarizeRun(t *testing.T) {
	for _, c := range []struct {
		state                                   runs.RunState
		passed, failed, unstable, open, logErrs int
		want                                    string
	}{
		{runs.RunRunning, 1, 1, 0, 0, 0, "running"},
		{runs.RunDone, 3, 1, 1, 1, 0, "failed"},
		{runs.RunUnknown, 3, 0, 1, 0, 0, "unstable"},
		{runs.RunUnknown, 3, 0, 0, 2, 0, "interrupted"},
		{runs.RunDone, 3, 0, 0, 0, 2, "passed"},
		{runs.RunUnknown, 0, 0, 0, 0, 2, "not_run"},
		{runs.RunDone, 0, 0, 0, 0, 0, "empty"},
	} {
		if got := summarizeRun(c.state, c.passed, c.failed, c.unstable, c.open, c.logErrs); got != c.want {
			t.Errorf("%+v: %s", c, got)
		}
	}
}

// The tests list shows each test's latest runs with their status, counting
// the tests themselves and not the runner protocols around them.
func TestTestsList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.App{}, &models.Test{}, &models.TestConfig{}, &models.TestConfigDevice{},
		&models.TestConfigUnity{}, &models.UnityTestFunction{}, &models.TestRun{}, &models.TestProtocol{},
		&models.TestRunLogEntry{}); err != nil {
		t.Fatal(err)
	}
	app := models.App{Name: "game"}
	db.Create(&app)
	smoke := models.Test{AppID: app.ID, Name: "smoke"}
	idle := models.Test{AppID: app.ID, Name: "never ran"}
	db.Create(&smoke)
	db.Create(&idle)
	db.Create(&models.Test{AppID: app.ID + 1, Name: "other app"})

	run := func(results ...models.TestResultState) uint {
		r := models.TestRun{TestID: smoke.ID}
		db.Create(&r)
		for _, res := range results {
			runner := models.TestProtocol{TestRunID: r.ID, TestName: "S/A", TestResult: res}
			db.Create(&runner)
			db.Create(&models.TestProtocol{TestRunID: r.ID, TestName: "S/A", TestResult: res, ParentTestProtocolID: &runner.ID})
		}
		return r.ID
	}
	failed := run(models.TestResultSuccess, models.TestResultFailed)
	notRun := run()
	db.Create(&models.TestRunLogEntry{TestRunID: notRun, Level: "error", Log: "no devices"})
	for i := 0; i < recentRunsPerTest; i++ {
		run(models.TestResultSuccess)
	}
	running := run(models.TestResultOpen)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db, runTracker: runs.NewTracker(nil)}
	s.runTracker.Started(running)
	r := gin.New()
	r.GET("/tests", func(c *gin.Context) { s.getTests(c, nil, &app) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/tests", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var list []testListEntry
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Name != "smoke" || list[1].Name != "never ran" {
		t.Fatalf("tests %+v", list)
	}
	if len(list[1].RecentRuns) != 0 {
		t.Errorf("a test that never ran has runs: %+v", list[1].RecentRuns)
	}
	recent := list[0].RecentRuns
	if len(recent) != recentRunsPerTest {
		t.Fatalf("%d runs, want the latest %d", len(recent), recentRunsPerTest)
	}
	first := recent[0]
	if first.ID != running || first.Status != "running" || first.State != runs.RunRunning || first.Open != 1 {
		t.Errorf("newest run %+v, want the running one with one open test (not two)", first)
	}
	if recent[1].Status != "passed" || recent[1].Passed != 1 {
		t.Errorf("second run %+v", recent[1])
	}
	for _, r := range recent {
		if r.ID == failed || r.ID == notRun {
			t.Errorf("run %d is older than the latest %d", r.ID, recentRunsPerTest)
		}
	}

	// with fewer runs the older ones show their status too
	db.Where("id > ?", notRun).Delete(&models.TestRun{})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/tests", nil))
	list = nil
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	statuses := map[uint]string{}
	for _, r := range list[0].RecentRuns {
		statuses[r.ID] = r.Status
	}
	if statuses[failed] != "failed" || statuses[notRun] != "not_run" {
		t.Errorf("statuses %v, want failed for %d and not_run for %d", statuses, failed, notRun)
	}
}
