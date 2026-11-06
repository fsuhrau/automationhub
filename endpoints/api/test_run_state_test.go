package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func runStateSetup(t *testing.T) (*Service, *gin.Engine) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Test{}, &models.TestConfig{}, &models.TestRun{}, &models.TestProtocol{},
		&models.TestRunLogEntry{}, &models.TestRunDeviceStatus{}, &models.ProtocolEntry{},
		&models.ProtocolPerformanceEntry{}, &models.AppBinary{}, &models.Device{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Test{Name: "smoke"})
	for i := 0; i < 3; i++ {
		db.Create(&models.TestRun{TestID: 1})
	}

	gin.SetMode(gin.TestMode)
	s := &Service{db: db, runTracker: runs.NewTracker(nil)}
	r := gin.New()
	r.GET("/run/:test_id/:run_id", func(c *gin.Context) { s.getTestRun(c, nil, nil) })
	r.GET("/last/:test_id", func(c *gin.Context) { s.getLastTestRun(c, nil, nil) })
	return s, r
}

func runResponse(t *testing.T, r http.Handler, path string) Response {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("%s: %d %s", path, w.Code, w.Body)
	}
	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

// The run page needs to know whether a run still executes, to offer Cancel
// and show progress; the run model has no status, so it comes from the tracker.
func TestRunResponseState(t *testing.T) {
	s, r := runStateSetup(t)
	s.runTracker.Started(2)
	s.runTracker.Started(3)
	s.runTracker.Finished(3, runs.RunResult{Success: true})

	for run, want := range map[int]runs.RunState{1: runs.RunUnknown, 2: runs.RunRunning, 3: runs.RunDone} {
		resp := runResponse(t, r, fmt.Sprintf("/run/1/%d", run))
		if resp.State != want {
			t.Errorf("run %d: state %q, want %q", run, resp.State, want)
		}
	}
	if resp := runResponse(t, r, "/last/1"); resp.TestRun.ID != 3 || resp.State != runs.RunDone || resp.PrevRunId != 2 {
		t.Errorf("last run: id %d state %q prev %d", resp.TestRun.ID, resp.State, resp.PrevRunId)
	}

	// without a tracker (e.g. a service built for other routes) runs aren't running
	s.runTracker = nil
	if resp := runResponse(t, r, "/run/1/2"); resp.State != runs.RunUnknown {
		t.Errorf("no tracker: state %q", resp.State)
	}
}
