package password

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRegister(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gdb := testkit.DB(t)
	p := New(gdb, []byte("test-secret-test-secret-test-sec"))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(p.Session("hub"))
	p.Routes(r)
	r.GET("/private", p.SessionHandler(), func(c *gin.Context) { c.String(http.StatusOK, "in") })
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

func TestLogin(t *testing.T) {
	r, _ := setupRegister(t)
	post(r, "/auth/register", `{"name":"Ada","email":"ada@example.com","password":"long enough"}`)

	for name, c := range map[string]struct {
		body string
		code int
	}{
		"ok":             {`{"email":" ada@example.com ","password":"long enough"}`, http.StatusOK},
		"wrong password": {`{"email":"ada@example.com","password":"nope nope"}`, http.StatusUnauthorized},
		"unknown email":  {`{"email":"bob@example.com","password":"long enough"}`, http.StatusUnauthorized},
		"bad json":       {`{"email":""}`, http.StatusBadRequest},
	} {
		if w := post(r, "/auth/login", c.body); w.Code != c.code {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
}

func TestSessionGuardAndLogout(t *testing.T) {
	r, _ := setupRegister(t)
	do := func(method, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		r.ServeHTTP(w, req)
		return w
	}

	if w := do("GET", "/private", nil); w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), `"url":"/login"`) {
		t.Errorf("guard without session: %d %s", w.Code, w.Body)
	}
	login := post(r, "/auth/register", `{"name":"Ada","email":"ada@example.com","password":"long enough"}`)
	cookies := login.Result().Cookies()
	if w := do("GET", "/private", cookies); w.Code != http.StatusOK {
		t.Errorf("guard with session: %d %s", w.Code, w.Body)
	}
	if w := do("GET", "/auth/session", cookies); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"user"`) {
		t.Errorf("session: %d %s", w.Code, w.Body)
	}
	out := do("POST", "/auth/logout", cookies)
	if out.Code != http.StatusOK {
		t.Fatalf("logout: %d", out.Code)
	}
	if w := do("GET", "/private", out.Result().Cookies()); w.Code != http.StatusUnauthorized {
		t.Errorf("after logout: %d", w.Code)
	}
}

func TestRegisterFailures(t *testing.T) {
	r, gdb := setupRegister(t)
	tooLong := strings.Repeat("x", 80)
	if w := post(r, "/auth/register", `{"name":"Bob","email":"bob@example.com","password":"`+tooLong+`"}`); w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "hash") {
		t.Errorf("unhashable password: %d %s", w.Code, w.Body)
	}
	if err := gdb.Migrator().DropTable(&models.UserAuth{}, &models.User{}); err != nil {
		t.Fatal(err)
	}
	if w := post(r, "/auth/register", `{"name":"Bob","email":"bob@example.com","password":"long enough"}`); w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "create user") {
		t.Errorf("store failure: %d %s", w.Code, w.Body)
	}
}

func TestRegistrationProblem(t *testing.T) {
	for name, c := range map[string]struct {
		r    registration
		want string
	}{
		"fine":    {registration{"a@b.c", "12345678", "A"}, ""},
		"empty":   {registration{"", "12345678", "A"}, "required"},
		"no at":   {registration{"abc", "12345678", "A"}, "email"},
		"space":   {registration{"a @b.c", "12345678", "A"}, "email"},
		"short":   {registration{"a@b.c", "1234567", "A"}, "at least 8"},
		"unicode": {registration{"a@b.c", "äöüäöüäö", "A"}, ""},
	} {
		if got := c.r.problem(); !strings.Contains(got, c.want) || (c.want == "" && got != "") {
			t.Errorf("%s: %q", name, got)
		}
	}
}
