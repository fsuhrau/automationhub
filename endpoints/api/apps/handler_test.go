package apps_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/appfixture"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func newAppEnv(t *testing.T) *appfixture.Env { return appfixture.New(t) }

func TestAppsListIsScopedToTheProject(t *testing.T) {
	e := newAppEnv(t)
	e.DB.Create(&models.AppParameter{AppID: e.Mine.ID, Name: "level", Type: `{"type":"string"}`})
	rec := e.Do("GET", "/api/mine/apps", nil)
	got := testkit.Decode[[]models.App](t, rec)
	if rec.Code != http.StatusOK || len(got) != 1 || got[0].ID != e.Mine.ID || len(got[0].Parameter) != 1 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec := e.Do("GET", "/api/nope/apps", nil); rec.Code != http.StatusNotFound {
		t.Errorf("unknown project: %d", rec.Code)
	}
}

func TestCreateApp(t *testing.T) {
	e := newAppEnv(t)
	for body, code := range map[string]int{
		`{"name":"","platform":1,"identifier":"x"}`:     http.StatusBadRequest,
		`{"name":"Game","platform":1,"identifier":" "}`: http.StatusBadRequest,
		`{"name":"Game","platform":42}`:                 http.StatusBadRequest,
		`{`:                                             http.StatusBadRequest,
		`{"name":" Editor ","platform":6}`:              http.StatusOK,
	} {
		if rec := e.Raw("POST", "/api/mine/app", body); rec.Code != code {
			t.Errorf("create app %s: %d %s", body, rec.Code, rec.Body)
		}
	}
	var editor models.App
	if e.DB.First(&editor, "name = ?", "Editor").Error != nil || editor.ProjectID != e.Mine.ProjectID {
		t.Errorf("the editor app wasn't stored with a trimmed name in the project: %+v", editor)
	}
}

func TestUpdateAppRenamesOnlyWithinTheProject(t *testing.T) {
	e := newAppEnv(t)
	rec := e.Raw("PUT", e.AppPath("/"), `{"name":"Renamed","identifier":"ignored"}`)
	got := testkit.Decode[models.App](t, rec)
	if rec.Code != http.StatusOK || got.Name != "Renamed" || got.Identifier != e.Mine.Identifier {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec := e.Raw("PUT", e.AppPath("/"), `not json`); rec.Code != http.StatusBadRequest {
		t.Errorf("not JSON: %d", rec.Code)
	}
	foreign := fmt.Sprintf("/api/mine/app/%d/", e.Other.ID)
	if rec := e.Raw("PUT", foreign, `{"name":"Hijacked"}`); rec.Code != http.StatusNotFound {
		t.Errorf("another project's app: %d", rec.Code)
	}
	var untouched models.App
	e.DB.First(&untouched, e.Other.ID)
	if untouched.Name == "Hijacked" {
		t.Error("another project's app was renamed")
	}
}

func TestAppParameters(t *testing.T) {
	e := newAppEnv(t)
	path := e.AppPath("/parameter")
	rec := e.Raw("POST", path, `{"name":" Market ","type":{"type":"option","defaultValue":"de","options":[" de","en","de",""]}}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"options":["de","en"]`) || !strings.Contains(rec.Body.String(), `"name":"Market"`) {
		t.Fatalf("add: %d %s", rec.Code, rec.Body)
	}
	var market models.AppParameter
	e.DB.First(&market, "name = ?", "Market")
	if rec := e.Raw("POST", path, `{"name":"Proxy","type":{"type":"string","defaultValue":"http://x"}}`); rec.Code != http.StatusOK {
		t.Fatalf("add text: %d %s", rec.Code, rec.Body)
	}

	for name, body := range map[string]string{
		"not JSON":           `{`,
		"no name":            `{"name":" ","type":{"type":"string"}}`,
		"taken name":         `{"name":"market","type":{"type":"string"}}`,
		"no options":         `{"name":"X","type":{"type":"option","options":[" "]}}`,
		"default not option": `{"name":"X","type":{"type":"option","defaultValue":"fr","options":["de"]}}`,
		"unknown type":       `{"name":"X","type":{"type":"number"}}`,
	} {
		if rec := e.Raw("POST", path, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}

	own := fmt.Sprintf("%s/%d", path, market.ID)
	rec = e.Raw("PUT", own, `{"name":"Market","type":{"type":"option","defaultValue":"en","options":["de","en","fr"]}}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"defaultValue":"en"`) || strings.Contains(rec.Body.String(), `"identifier"`) {
		t.Errorf("update: %d %s", rec.Code, rec.Body)
	}
	if rec := e.Raw("PUT", own, `{"name":"Proxy","type":{"type":"string"}}`); rec.Code != http.StatusBadRequest {
		t.Errorf("rename onto another parameter: %d", rec.Code)
	}
	if rec := e.Raw("PUT", own, `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("update with a broken body: %d", rec.Code)
	}
	if rec := e.Raw("PUT", path+"/999", `{"name":"x","type":{"type":"string"}}`); rec.Code != http.StatusNotFound {
		t.Errorf("update unknown: %d", rec.Code)
	}
	if rec := e.Raw("PUT", path+"/abc", `{}`); rec.Code != http.StatusBadRequest {
		t.Errorf("update with a non-numeric id: %d", rec.Code)
	}
}

func TestRemoveAppParameter(t *testing.T) {
	e := newAppEnv(t)
	p := models.AppParameter{AppID: e.Mine.ID, Name: "level", Type: `{"type":"string"}`}
	e.DB.Create(&p)
	foreign := fmt.Sprintf("/api/other/app/%d/parameter/%d", e.Other.ID, p.ID)
	if rec := e.Do("DELETE", foreign, nil); rec.Code != http.StatusOK {
		t.Fatalf("another app's delete: %d", rec.Code)
	}
	var n int64
	e.DB.Model(&models.AppParameter{}).Count(&n)
	if n != 1 {
		t.Fatalf("%d parameters, want the one: another app must not delete it", n)
	}
	path := fmt.Sprintf("%s/%d", e.AppPath("/parameter"), p.ID)
	if rec := e.Do("DELETE", path, nil); rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	e.DB.Model(&models.AppParameter{}).Count(&n)
	if n != 0 {
		t.Errorf("%d parameters left", n)
	}
	if rec := e.Do("DELETE", e.AppPath("/parameter/abc"), nil); rec.Code != http.StatusBadRequest {
		t.Errorf("non-numeric id: %d", rec.Code)
	}
}

// failWrites makes the database refuse every update and delete of the table.
func failWrites(t *testing.T, e *appfixture.Env, table string) {
	t.Helper()
	for _, op := range []string{"UPDATE", "DELETE"} {
		trigger := fmt.Sprintf("CREATE TRIGGER no_%s_%s BEFORE %s ON %s BEGIN SELECT RAISE(ABORT, 'read only'); END",
			strings.ToLower(op), table, op, table)
		if err := e.DB.Exec(trigger).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestStoreFailuresAreServerErrors(t *testing.T) {
	e := newAppEnv(t)
	p := models.AppParameter{AppID: e.Mine.ID, Name: "level", Type: `{"type":"string"}`}
	e.DB.Create(&p)
	failWrites(t, e, "apps")
	failWrites(t, e, "app_parameters")

	if rec := e.Raw("PUT", e.AppPath("/"), `{"name":"x"}`); rec.Code != http.StatusInternalServerError {
		t.Errorf("update app: %d %s", rec.Code, rec.Body)
	}
	if rec := e.Do("DELETE", fmt.Sprintf("%s/%d", e.AppPath("/parameter"), p.ID), nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("remove parameter: %d %s", rec.Code, rec.Body)
	}
	if err := e.DB.Migrator().DropTable(&models.AppParameter{}); err != nil {
		t.Fatal(err)
	}
	if rec := e.Do("GET", "/api/mine/apps", nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("list apps: %d %s", rec.Code, rec.Body)
	}
}
