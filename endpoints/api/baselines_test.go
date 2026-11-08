package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// A protocol's performance can be compared with another run of the same test
// picked by bundle name, version, tag or run id, on the same device or on
// any; runs without performance, other tests and other apps aren't offered.
func TestProtocolBaselines(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.App{}, &models.Test{}, &models.TestRun{}, &models.TestProtocol{},
		&models.ProtocolPerformanceEntry{}, &models.AppBinary{}, &models.Device{}); err != nil {
		t.Fatal(err)
	}
	app := models.App{Name: "game"}
	other := models.App{Name: "other"}
	db.Create(&app)
	db.Create(&other)
	test := models.Test{AppID: app.ID, Name: "perf"}
	foreign := models.Test{AppID: other.ID, Name: "perf"}
	db.Create(&test)
	db.Create(&foreign)
	phone := models.Device{Name: "phone"}
	tablet := models.Device{Name: "tablet", Alias: "Tab"}
	db.Create(&phone)
	db.Create(&tablet)

	bundle := func(version, tags string) models.AppBinary {
		b := models.AppBinary{AppID: app.ID, Name: "Game", Version: version, Tags: tags}
		db.Create(&b)
		return b
	}
	release := bundle("1.0", "release,store")
	candidate := bundle("1.1", "release_candidate")
	nightly := bundle("1.2", "nightly")

	// protocol adds a run of the test with one protocol of "Suite/Perf"
	protocol := func(test models.Test, b models.AppBinary, device models.Device, fps float64, perf bool) models.TestProtocol {
		r := models.TestRun{TestID: test.ID, AppBinaryID: b.ID}
		db.Create(&r)
		p := models.TestProtocol{TestRunID: r.ID, DeviceID: &device.ID, TestName: "Suite/Perf", AvgFPS: fps}
		db.Create(&p)
		if perf {
			db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1, FPS: fps})
			db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "end", Runtime: 3, FPS: fps})
		}
		return p
	}
	onRelease := protocol(test, release, phone, 60, true)
	onCandidate := protocol(test, candidate, phone, 55, true)
	onTablet := protocol(test, release, tablet, 30, true)
	unmeasured := protocol(test, release, phone, 0, false)
	inOtherApp := protocol(foreign, release, phone, 60, true)
	current := protocol(test, nightly, phone, 40, true)
	later := protocol(test, nightly, phone, 41, true)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.GET("/test/:test_id/run/:run_id/:protocol_id/baselines", func(c *gin.Context) { s.getProtocolBaselines(c, nil, &app) })
	r.GET("/test/:test_id/run/:run_id/:protocol_id/baseline/:baseline_run_id", func(c *gin.Context) { s.getProtocolBaseline(c, nil, &app) })
	base := fmt.Sprintf("/test/%d/run/%d/%d", test.ID, current.TestRunID, current.ID)
	list := func(query string) []uint {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", base+"/baselines"+query, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", query, w.Code, w.Body.String())
		}
		var resp baselinesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		ids := []uint{}
		for _, c := range resp.Candidates {
			ids = append(ids, c.ProtocolID)
		}
		if query == "" {
			if fmt.Sprint(resp.Tags) != "[nightly release release_candidate store]" {
				t.Errorf("tags %v", resp.Tags)
			}
			if c := resp.Candidates[len(resp.Candidates)-1]; c.Bundle == nil || c.Bundle.Version != "1.0" || c.AvgFPS != 60 || c.Device != "phone" {
				t.Errorf("oldest candidate %+v", c)
			}
		}
		return ids
	}
	expect := func(query string, want ...models.TestProtocol) {
		t.Helper()
		ids := []uint{}
		for _, p := range want {
			ids = append(ids, p.ID)
		}
		if got := list(query); fmt.Sprint(got) != fmt.Sprint(ids) {
			t.Errorf("%q: got %v, want %v", query, got, ids)
		}
	}

	expect("", later, onCandidate, onRelease)
	expect("?anyDevice=true", later, onTablet, onCandidate, onRelease)
	expect("?tag=release", onRelease)
	expect("?tag=release_candidate", onCandidate)
	expect("?tag=release&anyDevice=true", onTablet, onRelease)
	expect("?q=1.1", onCandidate)
	expect("?q=CANDIDATE", onCandidate)
	expect(fmt.Sprintf("?q=%%23%d", onRelease.TestRunID), onRelease)
	expect("?q=1_1")
	expect("?limit=1", later)
	_ = unmeasured
	_ = inOtherApp

	get := func(path string) (int, baselineResponse) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		var resp baselineResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return w.Code, resp
	}
	code, b := get(fmt.Sprintf("%s/baseline/%d", base, onTablet.TestRunID))
	if code != http.StatusOK || b.ProtocolID != onTablet.ID || b.Device != "Tab" || len(b.Performance) != 2 || b.Bundle.Tags != "release,store" {
		t.Errorf("baseline: %d %+v", code, b)
	}
	if b.Performance[0].Checkpoint != "start" {
		t.Errorf("checkpoints out of order: %+v", b.Performance)
	}
	// a run with the test on several devices: the same device unless another is asked for
	both := models.TestRun{TestID: test.ID, AppBinaryID: release.ID}
	db.Create(&both)
	for _, d := range []models.Device{phone, tablet} {
		p := models.TestProtocol{TestRunID: both.ID, DeviceID: &d.ID, TestName: "Suite/Perf"}
		db.Create(&p)
		db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p.ID, Checkpoint: "start", Runtime: 1})
	}
	if _, b := get(fmt.Sprintf("%s/baseline/%d", base, both.ID)); b.DeviceID == nil || *b.DeviceID != phone.ID {
		t.Errorf("baseline on device %v, want the same device %d", b.DeviceID, phone.ID)
	}
	if _, b := get(fmt.Sprintf("%s/baseline/%d?deviceId=%d", base, both.ID, tablet.ID)); b.DeviceID == nil || *b.DeviceID != tablet.ID {
		t.Errorf("baseline on device %v, want the asked for %d", b.DeviceID, tablet.ID)
	}
	for _, p := range []models.TestProtocol{unmeasured, inOtherApp, current} {
		if code, _ := get(fmt.Sprintf("%s/baseline/%d", base, p.TestRunID)); code != http.StatusNotFound {
			t.Errorf("run %d offered as baseline: %d", p.TestRunID, code)
		}
	}
	// the protocol has to be in the run and the run in the test of the app
	for _, path := range []string{
		fmt.Sprintf("/test/%d/run/%d/%d/baselines", test.ID, onRelease.TestRunID, current.ID),
		fmt.Sprintf("/test/%d/run/%d/%d/baselines", foreign.ID, inOtherApp.TestRunID, inOtherApp.ID),
	} {
		if code, _ := get(path); code != http.StatusNotFound {
			t.Errorf("%s: %d", path, code)
		}
	}
}
