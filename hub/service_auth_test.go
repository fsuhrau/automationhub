package hub

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/config"
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
	u, err := url.Parse(oauth2.GetLoginURL("st"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("client_id") != "client-1" || q.Get("scope") != "read openid" || q.Get("redirect_uri") != "http://localhost/cb" {
		t.Fatalf("unexpected login url query: %v", q)
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
