package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func settingsSetup(t *testing.T) (*gin.Engine, *gorm.DB, *models.App) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.AppParameter{}, &models.AccessToken{}, &models.Node{}); err != nil {
		t.Fatal(err)
	}
	mine := models.Project{Identifier: "mine"}
	other := models.Project{Identifier: "other"}
	db.Create(&mine)
	db.Create(&other)
	app := models.App{ProjectID: mine.ID, Name: "game", Platform: models.PlatformTypeAndroid, Identifier: "com.game"}
	db.Create(&app)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	api := r.Group("/api")
	api.POST("/:project_id/settings/access_token", s.createAccessToken)
	api.DELETE("/:project_id/settings/access_token/:token_id", s.deleteAccessToken)
	api.POST("/:project_id/settings/node", s.createNode)
	api.DELETE("/:project_id/settings/nodes/:node_id", s.deleteNode)
	api.POST("/:project_id/settings/nodes/:node_id/token", s.issueNodeTokenHandler)
	api.POST("/:project_id/app", func(c *gin.Context) { s.createApp(c, &mine) })
	api.POST("/:project_id/app/:app_id/parameter", func(c *gin.Context) { s.addAppParameter(c, &mine, &app) })
	api.PUT("/:project_id/app/:app_id/parameter/:parameter_id", func(c *gin.Context) { s.updateAppParameter(c, &mine, &app) })
	return r, db, &app
}

func liveTokens(db *gorm.DB, where string, args ...interface{}) int64 {
	var n int64
	db.Model(&models.AccessToken{}).Where(where, args...).Count(&n)
	return n
}

// A project's settings only revoke the project's own tokens; before, any
// token could be deleted by id.
func TestAccessTokens(t *testing.T) {
	r, db, _ := settingsSetup(t)
	var other models.Project
	db.First(&other, "identifier = ?", "other")

	if w := send(r, "POST", "/api/mine/settings/access_token", `{"name":"  "}`); w.Code != http.StatusBadRequest {
		t.Errorf("no name: %d", w.Code)
	}
	w := send(r, "POST", "/api/mine/settings/access_token", `{"name":" ci "}`)
	var mine models.AccessToken
	if err := json.Unmarshal(w.Body.Bytes(), &mine); err != nil || w.Code != http.StatusCreated || mine.Name != "ci" || mine.Token == "" {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	foreign := models.AccessToken{ProjectID: &other.ID, Name: "theirs", Token: "t-other"}
	userID := uint(1)
	personal := models.AccessToken{UserID: &userID, Name: "personal", Token: "sha256:x"}
	db.Create(&foreign)
	db.Create(&personal)

	for _, id := range []uint{foreign.ID, personal.ID, 999} {
		if w := send(r, "DELETE", fmt.Sprintf("/api/mine/settings/access_token/%d", id), ""); w.Code != http.StatusNotFound {
			t.Errorf("deleting token %d of someone else: %d", id, w.Code)
		}
	}
	if liveTokens(db, "id IN ?", []uint{foreign.ID, personal.ID}) != 2 {
		t.Errorf("another project's or a personal token was deleted")
	}
	if w := send(r, "DELETE", fmt.Sprintf("/api/mine/settings/access_token/%d", mine.ID), ""); w.Code != http.StatusOK {
		t.Errorf("own token: %d", w.Code)
	}
	if liveTokens(db, "id = ?", mine.ID) != 0 {
		t.Errorf("own token still there")
	}
}

// A node is set up with its identifier and token: creating it returns the
// token once, a new token replaces the old one, deleting the node revokes it.
func TestNodes(t *testing.T) {
	r, db, _ := settingsSetup(t)
	if w := send(r, "POST", "/api/mine/settings/node", `{"name":""}`); w.Code != http.StatusBadRequest {
		t.Errorf("no name: %d", w.Code)
	}
	w := send(r, "POST", "/api/mine/settings/node", `{"name":" Mac mini "}`)
	var created newNode
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || w.Code != http.StatusCreated || created.Name != "Mac mini" || created.Identifier == "" || created.Token == "" {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	if liveTokens(db, "node_id = ? AND token = ?", created.ID, created.Token) != 1 {
		t.Fatalf("the returned token isn't the node's")
	}

	w = send(r, "POST", fmt.Sprintf("/api/mine/settings/nodes/%d/token", created.ID), "")
	var issued newNode
	if err := json.Unmarshal(w.Body.Bytes(), &issued); err != nil || w.Code != http.StatusOK || issued.Token == "" || issued.Token == created.Token {
		t.Fatalf("new token: %d %s", w.Code, w.Body)
	}
	if liveTokens(db, "node_id = ?", created.ID) != 1 || liveTokens(db, "token = ?", created.Token) != 0 {
		t.Errorf("the old token still works")
	}

	if w := send(r, "DELETE", fmt.Sprintf("/api/mine/settings/nodes/%d", created.ID), ""); w.Code != http.StatusOK {
		t.Fatalf("delete: %d", w.Code)
	}
	if liveTokens(db, "node_id = ?", created.ID) != 0 {
		t.Errorf("a deleted node's token still works")
	}
	if w := send(r, "DELETE", "/api/mine/settings/nodes/999", ""); w.Code != http.StatusNotFound {
		t.Errorf("unknown node: %d", w.Code)
	}
}

func TestAppsAndParameters(t *testing.T) {
	r, db, app := settingsSetup(t)

	for body, code := range map[string]int{
		`{"name":"","platform":1,"identifier":"x"}`:     http.StatusBadRequest,
		`{"name":"Game","platform":1,"identifier":" "}`: http.StatusBadRequest,
		`{"name":"Game","platform":42}`:                 http.StatusBadRequest,
		`{"name":" Editor ","platform":6}`:              http.StatusOK,
	} {
		if w := send(r, "POST", "/api/mine/app", body); w.Code != code {
			t.Errorf("create app %s: %d %s", body, w.Code, w.Body)
		}
	}
	var editor models.App
	if db.First(&editor, "name = ?", "Editor").Error != nil {
		t.Errorf("the editor app wasn't stored with a trimmed name")
	}

	path := fmt.Sprintf("/api/mine/app/%d/parameter", app.ID)
	w := send(r, "POST", path, `{"name":" Market ","type":{"type":"option","defaultValue":"de","options":[" de","en","de",""]}}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"options":["de","en"]`) || !strings.Contains(w.Body.String(), `"name":"Market"`) {
		t.Fatalf("add: %d %s", w.Code, w.Body)
	}
	var market models.AppParameter
	db.First(&market, "name = ?", "Market")
	if w := send(r, "POST", path, `{"name":"Proxy","type":{"type":"string","defaultValue":"http://x"}}`); w.Code != http.StatusOK {
		t.Fatalf("add text: %d %s", w.Code, w.Body)
	}

	for name, body := range map[string]string{
		"no name":            `{"name":" ","type":{"type":"string"}}`,
		"taken name":         `{"name":"market","type":{"type":"string"}}`,
		"no options":         `{"name":"X","type":{"type":"option","options":[" "]}}`,
		"default not option": `{"name":"X","type":{"type":"option","defaultValue":"fr","options":["de"]}}`,
		"unknown type":       `{"name":"X","type":{"type":"number"}}`,
	} {
		if w := send(r, "POST", path, body); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}

	// an update keeps its own name, returns the parameter, and can't take another's name
	w = send(r, "PUT", fmt.Sprintf("%s/%d", path, market.ID), `{"name":"Market","type":{"type":"option","defaultValue":"en","options":["de","en","fr"]}}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"defaultValue":"en"`) || strings.Contains(w.Body.String(), `"identifier"`) {
		t.Errorf("update: %d %s", w.Code, w.Body)
	}
	if w := send(r, "PUT", fmt.Sprintf("%s/%d", path, market.ID), `{"name":"Proxy","type":{"type":"string"}}`); w.Code != http.StatusBadRequest {
		t.Errorf("rename onto another parameter: %d", w.Code)
	}
}
