package password

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupRegister(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&models.User{}, &models.UserAuth{}); err != nil {
		t.Fatal(err)
	}
	Setup(gdb, []byte("test-secret-test-secret-test-sec"))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Session("hub"))
	Routes(r)
	return r, gdb
}

func post(r http.Handler, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// Registering says what's wrong instead of failing: a short password, an
// address that isn't one, a taken email, or a taken name (names are unique,
// which used to end in "Failed to create user").
func TestRegister(t *testing.T) {
	r, gdb := setupRegister(t)

	w := post(r, "/auth/register", `{"name":" Ada ","email":"ada@example.com","password":"long enough"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"name":"Ada"`) {
		t.Fatalf("register: %d %s", w.Code, w.Body)
	}

	for name, c := range map[string]struct {
		body string
		code int
		says string
	}{
		"short password": {`{"name":"Bob","email":"bob@example.com","password":"short"}`, http.StatusBadRequest, "at least 8"},
		"no address":     {`{"name":"Bob","email":"bob","password":"long enough"}`, http.StatusBadRequest, "email address"},
		"taken email":    {`{"name":"Bob","email":"ada@example.com","password":"long enough"}`, http.StatusConflict, "exists already"},
		"taken name":     {`{"name":"Ada","email":"other@example.com","password":"long enough"}`, http.StatusConflict, "has this name"},
		"missing fields": {`{"name":"","email":"","password":""}`, http.StatusBadRequest, ""},
	} {
		w := post(r, "/auth/register", c.body)
		if w.Code != c.code || !strings.Contains(w.Body.String(), c.says) {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	var users int64
	gdb.Model(&models.User{}).Count(&users)
	if users != 1 {
		t.Errorf("%d users, want only Ada", users)
	}

	// the new account can sign in, and a signed out session says how to sign in
	if w := post(r, "/auth/login", `{"email":"ada@example.com","password":"long enough"}`); w.Code != http.StatusOK {
		t.Errorf("login: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/auth/session", nil))
	if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), `"method":"password"`) {
		t.Errorf("session: %d %s", w.Code, w.Body)
	}
}
