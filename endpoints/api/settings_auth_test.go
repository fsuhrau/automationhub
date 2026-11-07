package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// The settings hand out tokens, so a project token must not reach them:
// before, it could create a node and get a node token that works in every
// project.
func TestSettingsRefuseAccessTokens(t *testing.T) {
	_, db, _ := settingsSetup(t)
	var mine models.Project
	db.First(&mine, "identifier = ?", "mine")
	node := models.Node{Identifier: "n1", Name: "node"}
	db.Create(&node)
	db.Create(&models.AccessToken{ProjectID: &mine.ID, Name: "ci", Token: "project-token"})

	s := &Service{db: db}
	r := gin.New()
	session := func(c *gin.Context) { c.Next() } // a signed in browser
	s.registerSettings(r.Group("/api", token.SessionHandler(db, session)))

	do := func(method, path, tok string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(`{"name":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		if tok != "" {
			req.Header.Set("X-Auth-Token", tok)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	nodePath := fmt.Sprintf("/api/mine/settings/nodes/%d/token", node.ID)
	for _, c := range []struct{ method, path string }{
		{"POST", "/api/mine/settings/node"},
		{"POST", nodePath},
		{"GET", "/api/mine/settings/nodes"},
		{"POST", "/api/mine/settings/access_token"},
		{"GET", "/api/mine/settings/access_tokens"},
	} {
		if w := do(c.method, c.path, "project-token"); w.Code != http.StatusForbidden {
			t.Errorf("%s %s with a project token: expected 403, got %d %s", c.method, c.path, w.Code, w.Body.String())
		}
	}
	if n := liveTokens(db, "node_id IS NOT NULL"); n != 0 {
		t.Fatalf("a project token minted %d node tokens", n)
	}

	// the browser session still manages nodes
	if w := do("POST", "/api/mine/settings/node", ""); w.Code != http.StatusCreated || !strings.Contains(w.Body.String(), `"token"`) {
		t.Fatalf("session could not create a node: %d %s", w.Code, w.Body.String())
	}
}
