package projects_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// raw sends a body that isn't JSON built from a value.
func raw(env *apitest.Env, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return env.Request(req)
}

func TestListReturnsProjectsWithApps(t *testing.T) {
	env := apitest.New(t)
	testkit.App(env.DB)
	rec := env.Do("GET", "/api/projects", nil)
	got := testkit.Decode[[]models.Project](t, rec)
	if rec.Code != http.StatusOK || len(got) != 1 || len(got[0].Apps) != 1 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

// Creating a project only takes its name and gives it an identifier that
// is safe in URLs and not used before, not even by a deleted project.
func TestCreate(t *testing.T) {
	env := apitest.New(t)
	gone := models.Project{Name: "Game", Identifier: "game-2"}
	env.DB.Create(&gone)
	env.DB.Delete(&gone)
	create := func(body string) (int, models.Project) {
		rec := raw(env, "POST", "/api/project", body)
		var p models.Project
		if rec.Code == http.StatusCreated {
			p = testkit.Decode[models.Project](t, rec)
		}
		return rec.Code, p
	}

	if code, p := create(`{"name":" Game "}`); code != http.StatusCreated || p.Identifier != "game" || p.Name != "Game" {
		t.Errorf("first: %d %+v", code, p)
	}
	if _, p := create(`{"name":"game!"}`); p.Identifier != "game-3" {
		t.Errorf("second %q, want game-3 (game-2 belongs to a deleted project)", p.Identifier)
	}
	code, p := create(`{"name":"Sneaky","identifier":"admin","accessTokens":[{"token":"chosen"}],"apps":[{"name":"x"}]}`)
	if code != http.StatusCreated || p.Identifier != "sneaky" {
		t.Errorf("sneaky: %d %+v", code, p)
	}
	var tokens, apps int64
	env.DB.Model(&models.AccessToken{}).Count(&tokens)
	env.DB.Model(&models.App{}).Count(&apps)
	if tokens != 0 || apps != 0 {
		t.Errorf("a create stored %d tokens and %d apps from the request", tokens, apps)
	}
	for _, bad := range []string{`{"name":"  "}`, `name=x`} {
		if code, _ := create(bad); code != http.StatusBadRequest {
			t.Errorf("%s: %d", bad, code)
		}
	}
}

func TestRename(t *testing.T) {
	env := apitest.New(t)
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier, p.Name = "proj", "Project" })
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier, p.Name = "default_project", "Default" })

	rec := env.Do("PUT", "/api/project/proj", map[string]string{"name": " Renamed ", "identifier": "hijack"})
	got := testkit.Decode[models.Project](t, rec)
	if rec.Code != http.StatusOK || got.Name != "Renamed" || got.Identifier != "proj" {
		t.Fatalf("rename: %d %s", rec.Code, rec.Body)
	}
	rec = env.Do("PUT", "/api/project/default_project", map[string]string{"name": "Main", "identifier": "main"})
	if got := testkit.Decode[models.Project](t, rec); rec.Code != http.StatusOK || got.Identifier != "main" {
		t.Fatalf("default project: %d %s", rec.Code, rec.Body)
	}
}

func TestRenameFailures(t *testing.T) {
	env := apitest.New(t)
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "proj" })
	tests := []struct {
		name, path, body string
		want             int
	}{
		{"unknown project", "/api/project/nope", `{"name":"x"}`, http.StatusNotFound},
		{"blank name", "/api/project/proj", `{"name":" "}`, http.StatusBadRequest},
		{"malformed", "/api/project/proj", `{`, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := raw(env, "PUT", tc.path, tc.body); rec.Code != tc.want {
				t.Fatalf("%d %s", rec.Code, rec.Body)
			}
		})
	}
}

// Regression: the id went to gorm as a raw string, so a non-numeric id was
// a 500 ("no such column") and an unknown id a 500 as well.
func TestDelete(t *testing.T) {
	env := apitest.New(t)
	p := testkit.Project(env.DB)
	tests := []struct {
		name, path string
		want       int
	}{
		{"not a number", "/api/project/proj", http.StatusBadRequest},
		{"unknown id", "/api/project/99", http.StatusNotFound},
		{"known id", "/api/project/1", http.StatusOK},
		{"again", "/api/project/1", http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := env.Do("DELETE", tc.path, nil); rec.Code != tc.want {
				t.Fatalf("%d %s", rec.Code, rec.Body)
			}
		})
	}
	var n int64
	env.DB.Model(&models.Project{}).Where("id = ?", p.ID).Count(&n)
	if n != 0 {
		t.Fatal("the project is still there")
	}
}

func TestListFailsWhenTheStorageFails(t *testing.T) {
	env := apitest.New(t)
	if err := env.DB.Migrator().DropTable(&models.Project{}); err != nil {
		t.Fatal(err)
	}
	if rec := env.Do("GET", "/api/projects", nil); rec.Code != http.StatusInternalServerError {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
