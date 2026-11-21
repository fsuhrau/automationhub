package profile_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/endpoints/api/profile"
	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

const noUserMessage = "no signed in user; personal access tokens need a login (configure authentication on the master)"

// newEnv is an api with the users Alice and Bob, signed in as the user.
func newEnv(t *testing.T, opts ...apitest.Option) *apitest.Env {
	t.Helper()
	env := apitest.New(t, opts...)
	env.DB.Create(&models.User{Name: "Alice", Auth: []models.UserAuth{{Email: "alice@example.com", Provider: "password"}}})
	env.DB.Create(&models.User{Name: "Bob", Auth: []models.UserAuth{{Email: "bob@example.com", Provider: "github", ProviderUserID: "bob-login"}}})
	return env
}

func raw(env *apitest.Env, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return env.Request(req)
}

func TestProfile(t *testing.T) {
	tests := []struct {
		name string
		user authentication.User
		want profile.ProfileResponse
	}{
		{"by name", authentication.User{Name: "Alice", Login: "al", Email: "a@x"}, profile.ProfileResponse{ID: 1, Name: "Alice", Login: "al", Email: "a@x"}},
		// an OAuth2 session without a stored name resolves by its login
		{"by login", authentication.User{Name: "Unknown Name", Login: "bob-login"}, profile.ProfileResponse{ID: 2, Name: "Bob", Login: "bob-login"}},
		{"by email", authentication.User{Email: "alice@example.com"}, profile.ProfileResponse{ID: 1, Name: "Alice", Email: "alice@example.com"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newEnv(t, apitest.WithUser(tc.user))
			rec := env.Do("GET", "/api/me", nil)
			if got := testkit.Decode[profile.ProfileResponse](t, rec); rec.Code != http.StatusOK || got != tc.want {
				t.Fatalf("%d %+v", rec.Code, got)
			}
		})
	}
}

// Without authentication there is no user, so nothing of the profile works.
func TestSignedOut(t *testing.T) {
	env := newEnv(t)
	for _, path := range []string{"/api/me", "/api/me/tokens"} {
		rec := env.Do("GET", path, nil)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), noUserMessage) {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
	if rec := env.Do("POST", "/api/me/tokens", map[string]string{"name": "x"}); rec.Code != http.StatusNotFound {
		t.Errorf("create: %d", rec.Code)
	}
	if rec := env.Do("DELETE", "/api/me/tokens/1", nil); rec.Code != http.StatusNotFound {
		t.Errorf("delete: %d", rec.Code)
	}
}

func TestPersonalTokens(t *testing.T) {
	env := newEnv(t, apitest.WithUser(authentication.User{Name: "Alice"}))

	rec := env.Do("POST", "/api/me/tokens", map[string]string{"name": "claude"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	created := testkit.Decode[profile.CreatedPersonalToken](t, rec)
	if !token.IsPersonal(created.Token) || created.Hint != created.Token[len(created.Token)-4:] || created.Name != "claude" {
		t.Fatalf("created %+v", created)
	}
	var stored models.AccessToken
	env.DB.First(&stored, created.ID)
	if stored.Token != token.StoredPersonal(created.Token) || stored.UserID == nil {
		t.Fatalf("token stored in plain text or without user: %+v", stored)
	}

	// lists never contain the value
	rec = env.Do("GET", "/api/me/tokens", nil)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), created.Token) || strings.Contains(rec.Body.String(), stored.Token) || !strings.Contains(rec.Body.String(), `"name":"claude"`) {
		t.Fatalf("list: %d %s", rec.Code, rec.Body)
	}

	if rec := env.Do("DELETE", fmt.Sprintf("/api/me/tokens/%d", created.ID), nil); rec.Code != http.StatusNoContent {
		t.Fatalf("revoke: %d %s", rec.Code, rec.Body)
	}
	if rec := env.Do("GET", "/api/me/tokens", nil); rec.Body.String() != "[]" {
		t.Fatalf("revoked token still listed: %s", rec.Body)
	}
}

// Only the owner sees and revokes a personal token.
func TestPersonalTokensBelongToTheirUser(t *testing.T) {
	env := newEnv(t, apitest.WithUser(authentication.User{Name: "Bob"}))
	alice := uint(1)
	theirs := models.AccessToken{UserID: &alice, Name: "alices", Token: "sha256:a"}
	env.DB.Create(&theirs)

	if rec := env.Do("GET", "/api/me/tokens", nil); rec.Body.String() != "[]" {
		t.Fatalf("Bob sees Alice's tokens: %s", rec.Body)
	}
	if rec := env.Do("DELETE", fmt.Sprintf("/api/me/tokens/%d", theirs.ID), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("Bob revoked Alice's token: %d", rec.Code)
	}
}

func TestPersonalTokenRules(t *testing.T) {
	env := newEnv(t, apitest.WithUser(authentication.User{Name: "Alice"}))
	tests := []struct {
		name, body string
	}{
		{"empty name", `{"name":"  "}`},
		{"past expiry", `{"name":"x","expiresAt":"2001-01-01T00:00:00Z"}`},
		{"malformed", `{`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := raw(env, "POST", "/api/me/tokens", tc.body); rec.Code != http.StatusBadRequest {
				t.Fatalf("%d %s", rec.Code, rec.Body)
			}
		})
	}
	if rec := env.Do("DELETE", "/api/me/tokens/abc", nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("token id not a number: %d", rec.Code)
	}
}

// A leaked access token can't create new tokens or revoke the owner's.
func TestAccessTokensCannotManageTokens(t *testing.T) {
	env := newEnv(t, apitest.WithUser(authentication.User{Name: "Alice"}), apitest.WithAccessToken(1))
	if rec := env.Do("POST", "/api/me/tokens", map[string]string{"name": "x"}); rec.Code != http.StatusForbidden {
		t.Fatalf("token created with a token: %d", rec.Code)
	}
	if rec := env.Do("DELETE", "/api/me/tokens/1", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("token revoked with a token: %d", rec.Code)
	}
	if rec := env.Do("GET", "/api/me", nil); rec.Code != http.StatusOK {
		t.Fatalf("profile: %d", rec.Code)
	}
}
