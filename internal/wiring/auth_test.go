package wiring_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/internal/wiring"
)

func writeCredentials(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cred.json")
	if err := os.WriteFile(path, []byte(`{"clientid":"client-1","secret":"s3"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// masterWith builds a master on the in-memory database with the auth config.
func masterWith(t *testing.T, auth config.Auth) (*wiring.Master, error) {
	t.Helper()
	cfg := testConfig(t)
	cfg.Auth = auth
	return wiring.BuildMaster(cfg, wiring.WithDB(testkit.DB(t)), wiring.WithLogger(quietLogger()))
}

func TestMasterWithOAuth2UsesTheOAuth2Config(t *testing.T) {
	m, err := masterWith(t, config.Auth{OAuth2: &config.OAuth2{
		RedirectUrl: "http://localhost/cb",
		AuthUrl:     "http://idp/auth",
		TokenUrl:    "http://idp/token",
		Credentials: writeCredentials(t),
		Scopes:      []string{"read", "openid"},
		Secret:      "secret",
	}})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	m.Server.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/oauth2", nil))

	u, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Host != "idp" || q.Get("client_id") != "client-1" || q.Get("scope") != "read openid" || q.Get("redirect_uri") != "http://localhost/cb" || q.Get("state") == "" {
		t.Fatalf("unexpected login url: %v", u)
	}
}

func TestMasterWithGithubAndPasswordGuardsTheApi(t *testing.T) {
	cred := writeCredentials(t)
	tests := map[string]config.Auth{
		"github":   {Github: &config.Github{Credentials: cred, Scopes: []string{"user"}, Secret: "x"}},
		"password": {Password: &config.Password{Secret: "x"}},
	}
	for name, auth := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := masterWith(t, auth)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			m.Server.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/projects", nil))

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status %d", w.Code)
			}
		})
	}
}

func TestMasterWithoutAuthenticationIsOpen(t *testing.T) {
	m, err := masterWith(t, config.Auth{})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	m.Server.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/session", nil))
	open := httptest.NewRecorder()
	m.Server.Handler().ServeHTTP(open, httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	if w.Code != http.StatusOK || open.Code != http.StatusOK {
		t.Fatalf("session %d, projects %d", w.Code, open.Code)
	}
}

func TestMasterWithBrokenAuthenticationFailsToBuild(t *testing.T) {
	for name, auth := range map[string]config.Auth{
		"github": {Github: &config.Github{Credentials: "/does/not/exist"}},
		"oauth2": {OAuth2: &config.OAuth2{Credentials: "/does/not/exist"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := masterWith(t, auth); err == nil {
				t.Fatal("no error")
			}
		})
	}
}
