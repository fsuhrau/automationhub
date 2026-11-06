package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func testUpdateSetup(t *testing.T) (*gin.Engine, *gorm.DB, models.Test, models.Test, []models.Device) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.Test{}, &models.TestConfig{}, &models.TestConfigDevice{},
		&models.TestConfigUnity{}, &models.UnityTestFunction{}, &models.Device{}); err != nil {
		t.Fatal(err)
	}
	project := models.Project{Identifier: "p", CompanyID: 1}
	db.Create(&project)
	app := models.App{ProjectID: project.ID, Name: "game"}
	other := models.App{ProjectID: project.ID, Name: "other"}
	db.Create(&app)
	db.Create(&other)
	devices := []models.Device{{DeviceIdentifier: "a"}, {DeviceIdentifier: "b"}, {DeviceIdentifier: "foreign", CompanyID: 99}}
	for i := range devices {
		db.Create(&devices[i])
	}
	test := models.Test{AppID: app.ID, Name: "Smoke", TestConfig: models.TestConfig{
		Type: models.TestTypeUnity, Devices: []models.TestConfigDevice{{DeviceID: devices[0].ID}},
		Unity: &models.TestConfigUnity{UnityTestCategoryType: models.SelectedTestsOnly, UnityTestFunctions: []models.UnityTestFunction{
			{Assembly: "A", Class: "S", Method: "Old"},
		}},
	}}
	db.Create(&test)
	foreign := models.Test{AppID: other.ID, Name: "Other", TestConfig: models.TestConfig{Type: models.TestTypeUnity, Unity: &models.TestConfigUnity{}}}
	db.Create(&foreign)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.PUT("/test/:test_id", func(c *gin.Context) { s.updateTest(c, &project, &app) })
	return r, db, test, foreign, devices
}

func savedTest(t *testing.T, db *gorm.DB, id uint) (string, string, string) {
	t.Helper()
	var test models.Test
	db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, id)
	var devices, functions []string
	for _, d := range test.TestConfig.Devices {
		devices = append(devices, fmt.Sprint(d.DeviceID))
	}
	for _, f := range test.TestConfig.Unity.UnityTestFunctions {
		functions = append(functions, f.Class+"/"+f.Method)
	}
	sort.Strings(devices)
	sort.Strings(functions)
	return fmt.Sprintf("%s exec=%d all=%v", test.Name, test.TestConfig.ExecutionType, test.TestConfig.AllDevices),
		strings.Join(devices, ","),
		fmt.Sprintf("mode=%d cats=%q fns=%s", test.TestConfig.Unity.UnityTestCategoryType, test.TestConfig.Unity.Categories, strings.Join(functions, ","))
}

// Saving the test edit page replaces devices and test functions in one go,
// tidies the input, and refuses what can't run.
func TestUpdateTest(t *testing.T) {
	r, db, test, foreign, devices := testUpdateSetup(t)
	path := fmt.Sprintf("/test/%d", test.ID)

	body := fmt.Sprintf(`{"name":" Smoke 2 ","executionType":1,"allDevices":false,"devices":[%d,%d,%d],
		"unityTestCategoryType":2,"categories":" a, ,a ","testFunctions":[
		{"assembly":"A","class":"S","method":"New"},{"assembly":"A","class":"S","method":"New"},{"assembly":"A","class":"T","method":"Other"},{"method":" "}]}`,
		devices[1].ID, devices[0].ID, devices[1].ID)
	w := send(r, "PUT", path, body)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	head, devs, unity := savedTest(t, db, test.ID)
	if head != "Smoke 2 exec=1 all=false" || devs != fmt.Sprintf("%d,%d", devices[0].ID, devices[1].ID) ||
		unity != `mode=2 cats="a" fns=S/New,T/Other` {
		t.Errorf("saved %s | %s | %s", head, devs, unity)
	}
	var resp models.Test
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Name != "Smoke 2" || len(resp.TestConfig.Devices) != 2 {
		t.Errorf("response %s", w.Body)
	}

	// all devices drops the picked ones
	if w := send(r, "PUT", path, `{"name":"Smoke 2","executionType":0,"allDevices":true,"devices":[1],"unityTestCategoryType":0}`); w.Code != http.StatusOK {
		t.Fatalf("all devices: %d %s", w.Code, w.Body)
	}
	if head, devs, unity := savedTest(t, db, test.ID); head != "Smoke 2 exec=0 all=true" || devs != "" || unity != `mode=0 cats="" fns=` {
		t.Errorf("all devices saved %s | %s | %s", head, devs, unity)
	}

	before := fmt.Sprint(savedTest(t, db, test.ID))
	for name, c := range map[string]struct {
		body string
		code int
	}{
		"no name":            {`{"name":" ","allDevices":true}`, http.StatusBadRequest},
		"no devices":         {`{"name":"x","allDevices":false,"devices":[]}`, http.StatusBadRequest},
		"no tests picked":    {`{"name":"x","allDevices":true,"unityTestCategoryType":2,"testFunctions":[]}`, http.StatusBadRequest},
		"no category":        {`{"name":"x","allDevices":true,"unityTestCategoryType":1,"categories":" , "}`, http.StatusBadRequest},
		"unknown execution":  {`{"name":"x","allDevices":true,"executionType":7}`, http.StatusBadRequest},
		"unknown device":     {`{"name":"x","allDevices":false,"devices":[999]}`, http.StatusBadRequest},
		"another company's":  {fmt.Sprintf(`{"name":"x","allDevices":false,"devices":[%d]}`, devices[2].ID), http.StatusBadRequest},
		"not JSON":           {`name=x`, http.StatusBadRequest},
		"another app's test": {`{"name":"x","allDevices":true}`, http.StatusNotFound},
	} {
		p := path
		if name == "another app's test" {
			p = fmt.Sprintf("/test/%d", foreign.ID)
		}
		if w := send(r, "PUT", p, c.body); w.Code != c.code {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	if after := fmt.Sprint(savedTest(t, db, test.ID)); after != before {
		t.Errorf("a refused update changed the test: %s -> %s", before, after)
	}
}
