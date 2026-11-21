package hub

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func writeCredentials(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cred.json")
	if err := os.WriteFile(path, []byte(`{"clientid":"client-1","secret":"s3"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSetupAuthOAuth2OnlyUsesOAuth2Config(t *testing.T) {
	auth := config.Auth{OAuth2: &config.OAuth2{
		RedirectUrl: "http://localhost/cb",
		AuthUrl:     "http://idp/auth",
		TokenUrl:    "http://idp/token",
		Credentials: writeCredentials(t),
		Scopes:      []string{"read", "openid"},
		Secret:      "secret",
	}}

	sessionFunc, sessionHandler, routes := setupAuth(nil, auth)

	if sessionFunc == nil || sessionHandler == nil || routes == nil {
		t.Fatal("expected session middleware, handler and routes to be set")
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(sessionFunc)
	routes(r)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/oauth2", nil))
	u, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Host != "idp" || q.Get("client_id") != "client-1" || q.Get("scope") != "read openid" || q.Get("redirect_uri") != "http://localhost/cb" || q.Get("state") == "" {
		t.Fatalf("unexpected login url: %v", u)
	}
}

func TestSetupAuthProviders(t *testing.T) {
	cred := writeCredentials(t)
	tests := map[string]config.Auth{
		"github":   {Github: &config.Github{Credentials: cred, Scopes: []string{"user"}, Secret: "x"}},
		"password": {Password: &config.Password{Secret: "x"}},
	}
	for name, auth := range tests {
		t.Run(name, func(t *testing.T) {
			sessionFunc, sessionHandler, routes := setupAuth(nil, auth)
			if sessionFunc == nil || sessionHandler == nil || routes == nil {
				t.Fatal("expected session middleware, handler and routes to be set")
			}
		})
	}
}

func TestSetupAuthNothingConfigured(t *testing.T) {
	sessionFunc, sessionHandler, routes := setupAuth(nil, config.Auth{})
	if sessionFunc != nil || sessionHandler != nil || routes != nil {
		t.Fatal("expected nothing to be set up")
	}
}

func TestSetupAuthEndsTheProcessWhenTheProviderCannotBeBuilt(t *testing.T) {
	std := logrus.StandardLogger()
	defer func(exit func(int)) { std.ExitFunc = exit }(std.ExitFunc)
	std.ExitFunc = func(int) { panic("exit") }

	for name, auth := range map[string]config.Auth{
		"github": {Github: &config.Github{Credentials: "/does/not/exist"}},
		"oauth2": {OAuth2: &config.OAuth2{Credentials: "/does/not/exist"}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected the process to exit")
				}
			}()
			setupAuth(nil, auth)
		})
	}
}
