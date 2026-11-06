package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func profileSetup(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserAuth{}, &models.Project{}, &models.AccessToken{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.User{Name: "Alice", Auth: []models.UserAuth{{Email: "alice@example.com", Provider: "password"}}})
	db.Create(&models.User{Name: "Bob", Auth: []models.UserAuth{{Email: "bob@example.com", Provider: "github", ProviderUserID: "bob-login"}}})

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	// a fake session: X-Test-User names the user, X-Test-Token marks a
	// request that came in with an access token
	r.Use(func(c *gin.Context) {
		if name := c.GetHeader("X-Test-User"); name != "" {
			c.Set(authentication.UserKey, authentication.User{Name: name, Login: c.GetHeader("X-Test-Login")})
		}
		if c.GetHeader("X-Test-Token") != "" {
			c.Set(token.AuthTokenKey, uint(1))
		}
	})
	s.registerProfile(r.Group("/api"))
	return r, db
}

func call(r http.Handler, method, path, user, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if user != "" {
		req.Header.Set("X-Test-User", user)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPersonalTokens(t *testing.T) {
	r, db := profileSetup(t)

	w := call(r, "GET", "/api/me", "Alice", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"name":"Alice"`) {
		t.Fatalf("profile: %d %s", w.Code, w.Body)
	}

	w = call(r, "POST", "/api/me/tokens", "Alice", `{"name":"claude"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var created CreatedPersonalToken
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if !token.IsPersonal(created.Token) || created.Hint != created.Token[len(created.Token)-4:] || created.Name != "claude" {
		t.Fatalf("created %+v", created)
	}
	var stored models.AccessToken
	db.First(&stored, created.ID)
	if stored.Token != token.StoredPersonal(created.Token) || stored.UserID == nil {
		t.Fatalf("token stored in plain text or without user: %+v", stored)
	}

	// lists never contain the value
	w = call(r, "GET", "/api/me/tokens", "Alice", "")
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), created.Token) || strings.Contains(w.Body.String(), stored.Token) || !strings.Contains(w.Body.String(), `"name":"claude"`) {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	if w := call(r, "GET", "/api/me/tokens", "Bob", ""); w.Body.String() != "[]" {
		t.Fatalf("Bob sees Alice's tokens: %s", w.Body)
	}

	// only the owner can revoke, and only from a session
	path := "/api/me/tokens/" + jsonID(created.ID)
	if w := call(r, "DELETE", path, "Bob", ""); w.Code != http.StatusNotFound {
		t.Fatalf("Bob revoked Alice's token: %d", w.Code)
	}
	if w := call(r, "DELETE", path, "Alice", "", "X-Test-Token", "1"); w.Code != http.StatusForbidden {
		t.Fatalf("token revoked with a token: %d", w.Code)
	}
	if w := call(r, "DELETE", path, "Alice", ""); w.Code != http.StatusNoContent {
		t.Fatalf("revoke: %d %s", w.Code, w.Body)
	}
	if w := call(r, "GET", "/api/me/tokens", "Alice", ""); w.Body.String() != "[]" {
		t.Fatalf("revoked token still listed: %s", w.Body)
	}
}

func TestPersonalTokenRules(t *testing.T) {
	r, _ := profileSetup(t)
	if w := call(r, "POST", "/api/me/tokens", "Alice", `{"name":"x"}`, "X-Test-Token", "1"); w.Code != http.StatusForbidden {
		t.Fatalf("token created with a token: %d", w.Code)
	}
	if w := call(r, "POST", "/api/me/tokens", "Alice", `{"name":"  "}`); w.Code != http.StatusBadRequest {
		t.Fatalf("empty name: %d", w.Code)
	}
	if w := call(r, "POST", "/api/me/tokens", "Alice", `{"name":"x","expiresAt":"2001-01-01T00:00:00Z"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("past expiry: %d", w.Code)
	}
	// without authentication there is no user
	if w := call(r, "GET", "/api/me", "", ""); w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "login") {
		t.Fatalf("anonymous: %d %s", w.Code, w.Body)
	}
	// an OAuth2 session without a stored name resolves by its login
	if w := call(r, "GET", "/api/me", "Unknown Name", "", "X-Test-Login", "bob-login"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"name":"Bob"`) {
		t.Fatalf("login fallback: %d %s", w.Code, w.Body)
	}
}

func jsonID(id uint) string {
	b, _ := json.Marshal(id)
	return string(b)
}
