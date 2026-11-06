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

// The device page lists what ran on a device lately: the tests of this
// project, newest first, without the runner protocols that wrap them.
func TestDeviceProtocols(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.Test{}, &models.TestConfig{}, &models.TestRun{},
		&models.TestProtocol{}, &models.Device{}); err != nil {
		t.Fatal(err)
	}
	mine := models.Project{Identifier: "mine"}
	other := models.Project{Identifier: "other"}
	db.Create(&mine)
	db.Create(&other)
	device := models.Device{DeviceIdentifier: "phone"}
	elsewhere := models.Device{DeviceIdentifier: "tablet"}
	db.Create(&device)
	db.Create(&elsewhere)

	run := func(project models.Project) uint {
		app := models.App{ProjectID: project.ID, Name: "app"}
		db.Create(&app)
		test := models.Test{AppID: app.ID, Name: "smoke"}
		db.Create(&test)
		r := models.TestRun{TestID: test.ID}
		db.Create(&r)
		return r.ID
	}
	protocol := func(run uint, dev uint, name string, parent *uint) uint {
		p := models.TestProtocol{TestRunID: run, DeviceID: &dev, TestName: name, ParentTestProtocolID: parent}
		db.Create(&p)
		return p.ID
	}
	first, second, foreign := run(mine), run(mine), run(other)
	runner := protocol(first, device.ID, "S/A", nil)
	a := protocol(first, device.ID, "S/A", &runner)
	b := protocol(second, device.ID, "S/B", nil)  // a test without a runner
	protocol(second, elsewhere.ID, "S/C", nil)    // another device
	d := protocol(foreign, device.ID, "S/D", nil) // another project's test

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.GET("/:project/device/:device_id/protocols", func(c *gin.Context) {
		project := mine
		if c.Param("project") == "other" {
			project = other
		}
		s.getDeviceProtocols(c, &project)
	})
	get := func(path string) []models.TestProtocol {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		var list []models.TestProtocol
		if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
			t.Fatal(err)
		}
		return list
	}
	ids := func(list []models.TestProtocol) string {
		var out []uint
		for _, p := range list {
			out = append(out, p.ID)
		}
		return fmt.Sprint(out)
	}

	list := get(fmt.Sprintf("/mine/device/%d/protocols", device.ID))
	if got, want := ids(list), fmt.Sprint([]uint{b, a}); got != want {
		t.Errorf("protocols %s, want %s", got, want)
	}
	if len(list) > 0 && (list[0].TestRun == nil || list[0].TestRun.Test == nil || list[0].TestRun.Test.Name != "smoke") {
		t.Errorf("test run and test not loaded: %+v", list[0].TestRun)
	}
	if got := ids(get(fmt.Sprintf("/mine/device/%d/protocols?limit=1", device.ID))); got != fmt.Sprint([]uint{b}) {
		t.Errorf("limit 1: %s", got)
	}
	if got := ids(get(fmt.Sprintf("/other/device/%d/protocols", device.ID))); got != fmt.Sprint([]uint{d}) {
		t.Errorf("other project sees %s, want only its own %d", got, d)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/mine/device/999/protocols", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("unknown device: %d", w.Code)
	}
}
