package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// scopeFixture is two projects with one app, test and run each. The foreign
// project is created first, so its ids are the lower ones and an unscoped
// "earlier runs" query would pick them up.
type scopeFixture struct {
	router                      *gin.Engine
	db                          *gorm.DB
	mine, foreign               models.Test
	mineRun1, mineRun2, foreRun models.TestRun
	mineProto, foreProto        models.TestProtocol
	mineProto2                  models.TestProtocol
}

func newScopeFixture(t *testing.T) *scopeFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.Test{}, &models.TestConfig{}, &models.TestRun{},
		&models.TestProtocol{}, &models.Device{}, &models.TestRunDeviceStatus{}, &models.ProtocolPerformanceEntry{},
		&models.ProtocolEntry{}, &models.ProtocolKey{}, &models.TestRunLogEntry{}, &models.AppBinary{},
		&models.TestRunRecording{}); err != nil {
		t.Fatal(err)
	}
	f := &scopeFixture{db: db}
	device := models.Device{DeviceIdentifier: "d1"}
	db.Create(&device)

	foreignProject := models.Project{Identifier: "foreign"}
	db.Create(&foreignProject)
	foreignApp := models.App{ProjectID: foreignProject.ID, Name: "foreign"}
	db.Create(&foreignApp)
	f.foreign = models.Test{AppID: foreignApp.ID, Name: "foreign"}
	db.Create(&f.foreign)
	f.foreRun = models.TestRun{TestID: f.foreign.ID}
	db.Create(&f.foreRun)
	f.foreProto = f.protocol(f.foreRun.ID, device.ID, 100)
	db.Create(&models.TestRunDeviceStatus{TestRunID: f.foreRun.ID, DeviceID: device.ID, StartupTime: 1000})

	mineProject := models.Project{Identifier: "mine"}
	db.Create(&mineProject)
	mineApp := models.App{ProjectID: mineProject.ID, Name: "mine"}
	db.Create(&mineApp)
	f.mine = models.Test{AppID: mineApp.ID, Name: "smoke"}
	db.Create(&f.mine)
	f.mineRun1 = models.TestRun{TestID: f.mine.ID}
	db.Create(&f.mineRun1)
	f.mineProto = f.protocol(f.mineRun1.ID, device.ID, 10)
	db.Create(&models.TestRunDeviceStatus{TestRunID: f.mineRun1.ID, DeviceID: device.ID, StartupTime: 20})
	f.mineRun2 = models.TestRun{TestID: f.mine.ID}
	db.Create(&f.mineRun2)
	f.mineProto2 = f.protocol(f.mineRun2.ID, device.ID, 0)
	db.Create(&models.TestRunDeviceStatus{TestRunID: f.mineRun2.ID, DeviceID: device.ID, StartupTime: 30})

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	f.router = gin.New()
	f.router.GET("/test/:test_id/runs/last", func(c *gin.Context) { s.getLastTestRun(c, &mineProject, &mineApp) })
	f.router.GET("/test/:test_id/run/:run_id", func(c *gin.Context) { s.getTestRun(c, &mineProject, &mineApp) })
	f.router.GET("/test/:test_id/run/:run_id/:protocol_id", func(c *gin.Context) {
		s.getTestRunProtocol(c, &mineProject, &mineApp)
	})
	return f
}

// protocol stores a protocol of the shared device and test name that
// measured performance, so it counts for histories
func (f *scopeFixture) protocol(runID, deviceID uint, fps float64) models.TestProtocol {
	p := models.TestProtocol{TestRunID: runID, DeviceID: &deviceID, TestName: "S/A", AvgFPS: fps}
	f.db.Create(&p)
	f.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, FPS: fps})
	return p
}

func (f *scopeFixture) get(path string) (int, []byte) {
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w.Code, w.Body.Bytes()
}

// Runs and protocols are only reachable through a test of the app: a run of
// another project, or of another test, is not found.
func TestRunEndpointsRefuseOtherProjects(t *testing.T) {
	f := newScopeFixture(t)
	paths := map[string]string{
		"last run of a foreign test":         fmt.Sprintf("/test/%d/runs/last", f.foreign.ID),
		"foreign run":                        fmt.Sprintf("/test/%d/run/%d", f.foreign.ID, f.foreRun.ID),
		"foreign run through own test":       fmt.Sprintf("/test/%d/run/%d", f.mine.ID, f.foreRun.ID),
		"unknown run":                        fmt.Sprintf("/test/%d/run/9999", f.mine.ID),
		"foreign protocol":                   fmt.Sprintf("/test/%d/run/%d/%d", f.foreign.ID, f.foreRun.ID, f.foreProto.ID),
		"foreign protocol through own test":  fmt.Sprintf("/test/%d/run/%d/%d", f.mine.ID, f.foreRun.ID, f.foreProto.ID),
		"protocol of an unknown run":         fmt.Sprintf("/test/%d/run/9999/%d", f.mine.ID, f.mineProto.ID),
		"run of an unknown test":             "/test/9999/run/1",
		"last run of an unknown test":        "/test/9999/runs/last",
		"protocol of an unknown test's runs": "/test/9999/run/1/1",
	}
	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			code, body := f.get(path)
			if code != http.StatusNotFound {
				t.Errorf("%s: status %d, want 404: %s", path, code, body)
			}
		})
	}
}

// An id that is not a number never reaches the database.
func TestRunEndpointsRefuseNonNumericIDs(t *testing.T) {
	f := newScopeFixture(t)
	bad := url.PathEscape("1 OR 1=1")
	paths := map[string]string{
		"test id of last run":     "/test/" + bad + "/runs/last",
		"test id of run":          "/test/" + bad + "/run/1",
		"run id":                  fmt.Sprintf("/test/%d/run/%s", f.mine.ID, bad),
		"run id of protocol":      fmt.Sprintf("/test/%d/run/%s/1", f.mine.ID, bad),
		"protocol id of protocol": fmt.Sprintf("/test/%d/run/%d/%s", f.mine.ID, f.mineRun1.ID, bad),
		"negative run id":         fmt.Sprintf("/test/%d/run/-1", f.mine.ID),
	}
	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			code, body := f.get(path)
			if code != http.StatusBadRequest {
				t.Errorf("%s: status %d, want 400: %s", path, code, body)
			}
			var e ErrorResponse
			if json.Unmarshal(body, &e) != nil || e.Message == "" {
				t.Errorf("no JSON error message: %s", body)
			}
		})
	}
}

// The run of the own project answers with the same shape as before, and its
// prev/next and histories only use the own app's runs.
func TestOwnRunEndpoints(t *testing.T) {
	f := newScopeFixture(t)

	code, body := f.get(fmt.Sprintf("/test/%d/runs/last", f.mine.ID))
	var last Response
	if err := json.Unmarshal(body, &last); err != nil || code != http.StatusOK {
		t.Fatalf("last run: %d %s", code, body)
	}
	if last.TestRun.ID != f.mineRun2.ID || last.PrevRunId != f.mineRun1.ID || last.NextRunId != 0 {
		t.Errorf("last run %d prev %d next %d", last.TestRun.ID, last.PrevRunId, last.NextRunId)
	}

	code, body = f.get(fmt.Sprintf("/test/%d/run/%d", f.mine.ID, f.mineRun1.ID))
	var first Response
	if err := json.Unmarshal(body, &first); err != nil || code != http.StatusOK {
		t.Fatalf("run: %d %s", code, body)
	}
	if first.TestRun.ID != f.mineRun1.ID || first.PrevRunId != 0 || first.NextRunId != f.mineRun2.ID {
		t.Errorf("run %d prev %d next %d", first.TestRun.ID, first.PrevRunId, first.NextRunId)
	}
	if first.TestRun.Test == nil || len(first.TestRun.Protocols) != 1 || first.State == "" && first.Devices == nil {
		t.Errorf("run body %+v", first)
	}

	_, body = f.get(fmt.Sprintf("/test/%d/run/%d", f.mine.ID, f.mineRun2.ID))
	var second Response
	_ = json.Unmarshal(body, &second)
	if got := second.TestRun.DeviceStatus[0].HistAvgStartupTime; got != 20 {
		t.Errorf("startup history %d, want 20 (the foreign run's 1000 must not count)", got)
	}
}

func TestOwnProtocolHistoryIgnoresOtherProjects(t *testing.T) {
	f := newScopeFixture(t)
	code, body := f.get(fmt.Sprintf("/test/%d/run/%d/%d", f.mine.ID, f.mineRun2.ID, f.mineProto2.ID))
	var run models.TestRun
	if err := json.Unmarshal(body, &run); err != nil || code != http.StatusOK || len(run.Protocols) != 1 {
		t.Fatalf("protocol: %d %s", code, body)
	}
	p := run.Protocols[0]
	if p.HistAvgFPS != 10 || len(p.TestProtocolHistory) != 1 || p.TestProtocolHistory[0].ID != f.mineProto.ID {
		t.Errorf("history avg %v of %d protocols, want 10 of the own earlier one", p.HistAvgFPS, len(p.TestProtocolHistory))
	}
}

// /api/data/:name serves files of the data directory only.
func TestGetDataStaysInDataPath(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	if err := os.Mkdir(data, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(data, "shot.png"), []byte("png"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "secret.txt"), []byte("secret"), 0o644)
	old := apps.TestDataPath
	apps.TestDataPath = data
	t.Cleanup(func() { apps.TestDataPath = old })

	gin.SetMode(gin.TestMode)
	s := &Service{}
	serve := func(name string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Params = gin.Params{{Key: "name", Value: name}}
		s.getData(c)
		return w
	}
	if w := serve("shot.png"); w.Code != http.StatusOK || w.Body.String() != "png" {
		t.Errorf("own file: %d %q", w.Code, w.Body.String())
	}
	for _, name := range []string{"../secret.txt", "/../secret.txt", filepath.Join(root, "secret.txt"), "..", "", "a/../../secret.txt"} {
		w := serve(name)
		if w.Code != http.StatusBadRequest || w.Body.String() == "secret" {
			t.Errorf("%q: status %d %q, want 400", name, w.Code, w.Body.String())
		}
	}

	// through the router a traversal never matches the single-segment route
	r := gin.New()
	r.GET("/api/data/:name", s.getData)
	for _, path := range []string{"/api/data/..%2fsecret.txt", "/api/data/..%2f..%2fetc%2fpasswd", "/api/data/../secret.txt"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently || w.Body.String() == "secret" {
			t.Errorf("%s: status %d %q", path, w.Code, w.Body.String())
		}
	}
}

// A database failure is a 500, not a 404.
func TestRunEndpointsReportDatabaseErrors(t *testing.T) {
	for _, table := range []string{"tests", "test_runs"} {
		t.Run(table, func(t *testing.T) {
			f := newScopeFixture(t)
			if err := f.db.Migrator().DropTable(table); err != nil {
				t.Fatal(err)
			}
			code, body := f.get(fmt.Sprintf("/test/%d/run/%d", f.mine.ID, f.mineRun1.ID))
			if code != http.StatusInternalServerError {
				t.Errorf("status %d, want 500: %s", code, body)
			}
		})
	}
}
