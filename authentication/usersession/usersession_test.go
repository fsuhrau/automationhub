package usersession

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// failingSession keeps values in memory and fails to save them.
type failingSession struct {
	sessions.Session
	values map[interface{}]interface{}
}

func (s failingSession) Get(k interface{}) interface{} { return s.values[k] }
func (s failingSession) Set(k, v interface{})          { s.values[k] = v }
func (s failingSession) Delete(k interface{})          { delete(s.values, k) }
func (s failingSession) Save() error                   { return errors.New("save failed") }

func newRouter(opts Options) (*gin.Engine, *Guard) {
	gin.SetMode(gin.TestMode)
	g := New([]byte("test-secret-test-secret-test-sec"), opts)
	r := gin.New()
	r.Use(g.Middleware("hub"))
	g.Register(r)
	r.GET("/private", g.RequireUser(), func(c *gin.Context) { c.String(http.StatusOK, "in") })
	r.POST("/signin", func(c *gin.Context) {
		g.SignIn(c, authentication.User{Login: "ada", Name: "Ada"})
		c.String(http.StatusOK, "ok")
	})
	return r, g
}

func do(r http.Handler, method, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestSignedOutAnswers(t *testing.T) {
	tests := map[string]struct {
		opts       Options
		path, says string
		code       int
	}{
		"guard":              {Options{Method: "m", LoginURL: "/in"}, "/private", `"url":"/in"`, http.StatusUnauthorized},
		"status without url": {Options{Method: "m", LoginURL: "/in"}, "/auth/session", `{"error":"not signed in","method":"m"}`, http.StatusUnauthorized},
		"status with url":    {Options{Method: "m", LoginURL: "/in", StatusIncludesURL: true}, "/auth/session", `"url":"/in"`, http.StatusUnauthorized},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r, _ := newRouter(tc.opts)
			w := do(r, "GET", tc.path, nil)
			if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.says) {
				t.Errorf("%d %s", w.Code, w.Body)
			}
		})
	}
}

func TestSignInThenLogout(t *testing.T) {
	r, _ := newRouter(Options{Method: "m", LoginURL: "/in"})
	cookies := do(r, "POST", "/signin", nil).Result().Cookies()

	if w := do(r, "GET", "/private", cookies); w.Code != http.StatusOK {
		t.Errorf("guard: %d", w.Code)
	}
	if w := do(r, "GET", "/auth/session", cookies); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"login":"ada"`) {
		t.Errorf("status: %d %s", w.Code, w.Body)
	}
	out := do(r, "POST", "/auth/logout", cookies)
	if out.Code != http.StatusOK || out.Body.String() != "null" {
		t.Errorf("logout: %d %s", out.Code, out.Body)
	}
	if w := do(r, "GET", "/private", out.Result().Cookies()); w.Code != http.StatusUnauthorized {
		t.Errorf("after logout: %d", w.Code)
	}
	if w := do(r, "POST", "/auth/logout", nil); w.Code != http.StatusOK {
		t.Errorf("logout without session: %d", w.Code)
	}
}

func TestSaveFailuresAreLogged(t *testing.T) {
	_, g := newRouter(Options{Method: "m"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/", nil)
	ctx.Set(sessions.DefaultKey, failingSession{values: map[interface{}]interface{}{}})

	g.SignIn(ctx, authentication.User{Login: "ada"})
	if _, ok := g.User(ctx); !ok {
		t.Error("the user should be signed in for the request even if saving fails")
	}
	g.Logout(ctx)
	if _, ok := g.User(ctx); ok {
		t.Error("logout should remove the user")
	}
}
