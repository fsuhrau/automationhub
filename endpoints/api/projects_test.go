package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestProjectSlug(t *testing.T) {
	for name, want := range map[string]string{
		"Heroes of History": "heroes-of-history",
		"  R&D / Tests? ":   "r-d-tests",
		"Über Spiel 2":      "ber-spiel-2",
		"---":               "project",
		"日本":                "project",
		"a  --  b":          "a-b",
		"Trailing!!!":       "trailing",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa b": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	} {
		if got := projectSlug(name); got != want {
			t.Errorf("%q: %q, want %q", name, got, want)
		}
	}
}

// Creating a project only takes its name and gives it an identifier that
// is safe in URLs and not used before, not even by a deleted project.
func TestCreateProject(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.AccessToken{}, &models.App{}); err != nil {
		t.Fatal(err)
	}
	gone := models.Project{Name: "Game", Identifier: "game-2"}
	db.Create(&gone)
	db.Delete(&gone)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.POST("/project", s.createProject)
	create := func(body string) (int, models.Project) {
		w := send(r, "POST", "/project", body)
		var p models.Project
		_ = json.Unmarshal(w.Body.Bytes(), &p)
		return w.Code, p
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
	db.Model(&models.AccessToken{}).Count(&tokens)
	db.Model(&models.App{}).Count(&apps)
	if tokens != 0 || apps != 0 {
		t.Errorf("a create stored %d tokens and %d apps from the request", tokens, apps)
	}
	for _, bad := range []string{`{"name":"  "}`, `name=x`} {
		if code, _ := create(bad); code != http.StatusBadRequest {
			t.Errorf("%s: %d", bad, code)
		}
	}
}
