package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func deviceUpdateSetup(t *testing.T) (*gin.Engine, *gorm.DB, models.Device) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.Device{}, &models.CustomParameter{}, &models.DeviceParameter{},
		&models.ConnectionParameter{}); err != nil {
		t.Fatal(err)
	}
	device := models.Device{DeviceIdentifier: "phone", Alias: "Old alias", CustomParameter: []models.CustomParameter{
		{Key: "PIN", Value: "1234"}, {Key: "LANG", Value: "de"}, {Key: "GONE", Value: "x"},
	}}
	db.Create(&device)
	other := models.Device{DeviceIdentifier: "tablet", CustomParameter: []models.CustomParameter{{Key: "OTHER", Value: "o"}}}
	db.Create(&other)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	project := models.Project{}
	r := gin.New()
	r.PUT("/device/:device_id", func(c *gin.Context) { s.updateDevice(c, &project) })
	return r, db, device
}

func put(r http.Handler, id uint, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/device/%d", id), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func params(t *testing.T, db *gorm.DB, deviceID uint) string {
	t.Helper()
	var list []models.CustomParameter
	db.Where("device_id = ?", deviceID).Find(&list)
	var out []string
	for _, p := range list {
		out = append(out, p.Key+"="+p.Value)
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

// The edit page changes the alias and the custom parameters. Before, a
// renamed key deleted its parameter, edited values weren't saved (GORM only
// upserts the foreign key of existing associations), and an empty alias was
// ignored.
func TestUpdateDevice(t *testing.T) {
	r, db, device := deviceUpdateSetup(t)
	pin, lang := device.CustomParameter[0].ID, device.CustomParameter[1].ID
	var other models.CustomParameter
	db.Where("key = ?", "OTHER").First(&other)

	body := fmt.Sprintf(`{"alias":"  ","customParameter":[
		{"id":%d,"key":"PIN","value":"9999"},
		{"id":%d,"key":" LOCALE ","value":"en"},
		{"key":"NEW","value":"n"},
		{"id":%d,"key":"STOLEN","value":"s"}]}`, pin, lang, other.ID)
	w := put(r, device.ID, body)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if got, want := params(t, db, device.ID), "LOCALE=en NEW=n PIN=9999 STOLEN=s"; got != want {
		t.Errorf("parameters %q, want %q", got, want)
	}
	if got := params(t, db, 2); got != "OTHER=o" {
		t.Errorf("the other device's parameters changed: %q", got)
	}
	var saved models.Device
	db.First(&saved, device.ID)
	if saved.Alias != "" {
		t.Errorf("alias %q, want it cleared", saved.Alias)
	}
	var resp models.Device
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || len(resp.CustomParameter) != 4 {
		t.Errorf("response %s", w.Body)
	}
	var renamed models.CustomParameter
	db.First(&renamed, lang)
	if renamed.Key != "LOCALE" {
		t.Errorf("renamed parameter %+v, want the same row with the new key", renamed)
	}

	for name, bad := range map[string]string{
		"empty key":     `{"customParameter":[{"key":" ","value":"v"}]}`,
		"duplicate key": `{"customParameter":[{"key":"A","value":"1"},{"key":"A","value":"2"}]}`,
		"not JSON":      `alias=x`,
	} {
		if w := put(r, device.ID, bad); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	if got := params(t, db, device.ID); got != "LOCALE=en NEW=n PIN=9999 STOLEN=s" {
		t.Errorf("a rejected update changed the parameters: %q", got)
	}
	if w := put(r, 999, `{}`); w.Code != http.StatusNotFound {
		t.Errorf("unknown device: %d", w.Code)
	}
}
