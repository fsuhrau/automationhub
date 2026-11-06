package token

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserAuth{}, &models.Project{}, &models.AccessToken{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// server runs the middleware; the handler reports who it saw.
func server(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	session := func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	g := r.Group("/", SessionHandler(db, session))
	handler := func(c *gin.Context) {
		v, _ := c.Get(authentication.UserKey)
		switch u := v.(type) {
		case authentication.User:
			c.String(http.StatusOK, "user:"+u.Name+":"+u.Email)
		case string:
			c.String(http.StatusOK, "token:"+u)
		}
	}
	g.GET("/api/:project_id/x", handler)
	return r
}

func get(r http.Handler, path, header string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	if header != "" {
		req.Header.Set("X-Auth-Token", header)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestNewPersonal(t *testing.T) {
	plain, stored, hint, err := NewPersonal()
	if err != nil {
		t.Fatal(err)
	}
	if !IsPersonal(plain) || len(plain) < 40 || stored != StoredPersonal(plain) || stored == plain || plain[len(plain)-4:] != hint {
		t.Fatalf("unexpected token %q %q %q", plain, stored, hint)
	}
	other, _, _, _ := NewPersonal()
	if other == plain {
		t.Fatal("tokens repeat")
	}
}

func TestPersonalTokenActsAsUser(t *testing.T) {
	db := testDB(t)
	user := models.User{Name: "Fabian", Auth: []models.UserAuth{{Email: "f@example.com", Provider: "password"}}}
	db.Create(&user)
	db.Create(&models.Project{Name: "Game", Identifier: "game"})
	plain, stored, hint, _ := NewPersonal()
	tok := models.AccessToken{UserID: &user.ID, Name: "claude", Token: stored, Hint: hint}
	db.Create(&tok)
	r := server(db)

	w := get(r, "/api/game/x", plain)
	if w.Code != http.StatusOK || w.Body.String() != "user:Fabian:f@example.com" {
		t.Fatalf("personal token: %d %s", w.Code, w.Body)
	}
	// works in any project, like the session of the user
	if w := get(r, "/api/other/x", plain); w.Code != http.StatusOK {
		t.Fatalf("other project: %d %s", w.Code, w.Body)
	}
	var reloaded models.AccessToken
	db.First(&reloaded, tok.ID)
	if reloaded.LastUsedAt == nil {
		t.Fatal("last used not recorded")
	}

	// the stored hash is not a token
	if w := get(r, "/api/game/x", stored); w.Code != http.StatusUnauthorized {
		t.Fatalf("hash accepted: %d", w.Code)
	}
	if w := get(r, "/api/game/x", PersonalPrefix+"wrong"); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token accepted: %d", w.Code)
	}

	// revoked (soft deleted) and expired tokens are refused
	db.Delete(&tok)
	if w := get(r, "/api/game/x", plain); w.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token accepted: %d", w.Code)
	}
	plain2, stored2, _, _ := NewPersonal()
	past := time.Now().Add(-time.Hour)
	db.Create(&models.AccessToken{UserID: &user.ID, Name: "old", Token: stored2, ExpiresAt: &past})
	if w := get(r, "/api/game/x", plain2); w.Code != http.StatusUnauthorized {
		t.Fatalf("expired token accepted: %d", w.Code)
	}
}

func TestProjectTokenUnchanged(t *testing.T) {
	db := testDB(t)
	project := models.Project{Name: "Game", Identifier: "game"}
	db.Create(&project)
	db.Create(&models.AccessToken{ProjectID: &project.ID, Name: "ci", Token: "3f1c-uuid"})
	r := server(db)

	if w := get(r, "/api/game/x", "3f1c-uuid"); w.Code != http.StatusOK || w.Body.String() != "token:ci" {
		t.Fatalf("project token: %d %s", w.Code, w.Body)
	}
	if w := get(r, "/api/other/x", "3f1c-uuid"); w.Code != http.StatusBadRequest {
		t.Fatalf("project token in another project: %d", w.Code)
	}
	if w := get(r, "/api/game/x", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token falls through to the session handler: %d", w.Code)
	}
}
