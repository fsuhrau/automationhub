package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// The test page pages through the runs of a test, newest first, with what
// each ran with and when it ended; a test of another app is not found.
func TestTestRunsPage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.App{}, &models.Test{}, &models.TestRun{}, &models.TestProtocol{},
		&models.TestRunLogEntry{}, &models.AppBinary{}); err != nil {
		t.Fatal(err)
	}
	app := models.App{Name: "game"}
	other := models.App{Name: "other"}
	db.Create(&app)
	db.Create(&other)
	test := models.Test{AppID: app.ID, Name: "smoke"}
	foreign := models.Test{AppID: other.ID, Name: "foreign"}
	db.Create(&test)
	db.Create(&foreign)
	bundle := models.AppBinary{AppID: app.ID, Name: "Game", Version: "1.2"}
	db.Create(&bundle)

	var ids []uint
	for i := 0; i < 5; i++ {
		r := models.TestRun{TestID: test.ID, AppBinaryID: bundle.ID, Parameter: fmt.Sprintf("Run=%d", i)}
		db.Create(&r)
		ids = append(ids, r.ID)
	}
	end := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	later := end.Add(time.Minute)
	newest := ids[4]
	db.Create(&models.TestProtocol{TestRunID: newest, TestName: "S/A", TestResult: models.TestResultSuccess, EndedAt: &end})
	db.Create(&models.TestProtocol{TestRunID: newest, TestName: "S/B", TestResult: models.TestResultFailed, EndedAt: &later})

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.GET("/test/:test_id/runs", func(c *gin.Context) { s.getTestRuns(c, nil, &app) })
	get := func(path string) (int, testRunsPage) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		var page testRunsPage
		_ = json.Unmarshal(w.Body.Bytes(), &page)
		return w.Code, page
	}

	code, page := get(fmt.Sprintf("/test/%d/runs?limit=2", test.ID))
	if code != http.StatusOK || len(page.Runs) != 2 || !page.More {
		t.Fatalf("first page: %d %+v", code, page)
	}
	first := page.Runs[0]
	if first.ID != newest || first.Status != "failed" || first.Passed != 1 || first.Failed != 1 {
		t.Errorf("newest run %+v", first.runSummary)
	}
	if first.Bundle == nil || first.Bundle.Version != "1.2" || first.Parameter != "Run=4" {
		t.Errorf("newest run ran with %+v %q", first.Bundle, first.Parameter)
	}
	if first.EndedAt == nil || !first.EndedAt.Equal(later) {
		t.Errorf("ended at %v, want the last protocol's end %v", first.EndedAt, later)
	}
	if page.Runs[1].EndedAt != nil || page.Runs[1].Status != "empty" {
		t.Errorf("a run without protocols: %+v", page.Runs[1])
	}

	_, older := get(fmt.Sprintf("/test/%d/runs?limit=2&before=%d", test.ID, page.Runs[1].ID))
	if len(older.Runs) != 2 || older.Runs[0].ID != ids[2] || !older.More {
		t.Errorf("second page %+v", older)
	}
	_, last := get(fmt.Sprintf("/test/%d/runs?limit=2&before=%d", test.ID, ids[1]))
	if len(last.Runs) != 1 || last.Runs[0].ID != ids[0] || last.More {
		t.Errorf("last page %+v", last)
	}
	if _, all := get(fmt.Sprintf("/test/%d/runs", test.ID)); len(all.Runs) != 5 || all.More {
		t.Errorf("default page: %d runs, more %v", len(all.Runs), all.More)
	}

	if code, _ := get(fmt.Sprintf("/test/%d/runs", foreign.ID)); code != http.StatusNotFound {
		t.Errorf("another app's test: %d", code)
	}
}
