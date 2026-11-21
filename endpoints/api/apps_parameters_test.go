package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestAppsAndParameters(t *testing.T) {
	env := apitest.New(t)
	mine := testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "mine" })
	app := testkit.App(env.DB, func(a *models.App) {
		a.ProjectID, a.Name, a.Platform, a.Identifier = mine.ID, "game", models.PlatformTypeAndroid, "com.game"
	})

	for body, code := range map[string]int{
		`{"name":"","platform":1,"identifier":"x"}`:     http.StatusBadRequest,
		`{"name":"Game","platform":1,"identifier":" "}`: http.StatusBadRequest,
		`{"name":"Game","platform":42}`:                 http.StatusBadRequest,
		`{"name":" Editor ","platform":6}`:              http.StatusOK,
	} {
		if w := send(env, "POST", "/api/mine/app", body); w.Code != code {
			t.Errorf("create app %s: %d %s", body, w.Code, w.Body)
		}
	}
	var editor models.App
	if env.DB.First(&editor, "name = ?", "Editor").Error != nil {
		t.Errorf("the editor app wasn't stored with a trimmed name")
	}

	path := fmt.Sprintf("/api/mine/app/%d/parameter", app.ID)
	w := send(env, "POST", path, `{"name":" Market ","type":{"type":"option","defaultValue":"de","options":[" de","en","de",""]}}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"options":["de","en"]`) || !strings.Contains(w.Body.String(), `"name":"Market"`) {
		t.Fatalf("add: %d %s", w.Code, w.Body)
	}
	var market models.AppParameter
	env.DB.First(&market, "name = ?", "Market")
	if w := send(env, "POST", path, `{"name":"Proxy","type":{"type":"string","defaultValue":"http://x"}}`); w.Code != http.StatusOK {
		t.Fatalf("add text: %d %s", w.Code, w.Body)
	}

	for name, body := range map[string]string{
		"no name":            `{"name":" ","type":{"type":"string"}}`,
		"taken name":         `{"name":"market","type":{"type":"string"}}`,
		"no options":         `{"name":"X","type":{"type":"option","options":[" "]}}`,
		"default not option": `{"name":"X","type":{"type":"option","defaultValue":"fr","options":["de"]}}`,
		"unknown type":       `{"name":"X","type":{"type":"number"}}`,
	} {
		if w := send(env, "POST", path, body); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}

	// an update keeps its own name, returns the parameter, and can't take another's name
	w = send(env, "PUT", fmt.Sprintf("%s/%d", path, market.ID), `{"name":"Market","type":{"type":"option","defaultValue":"en","options":["de","en","fr"]}}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"defaultValue":"en"`) || strings.Contains(w.Body.String(), `"identifier"`) {
		t.Errorf("update: %d %s", w.Code, w.Body)
	}
	if w := send(env, "PUT", fmt.Sprintf("%s/%d", path, market.ID), `{"name":"Proxy","type":{"type":"string"}}`); w.Code != http.StatusBadRequest {
		t.Errorf("rename onto another parameter: %d", w.Code)
	}
}

// send sends a raw JSON body to the api.
func send(env *apitest.Env, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return env.Request(req)
}
