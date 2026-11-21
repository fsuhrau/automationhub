// Package oauth2 signs users in with any OAuth2 provider that has a user endpoint.
package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// ProviderName is stored with these users and reported as sign in method.
const ProviderName = "oauth2"

// Endpoints are the URLs of the OAuth2 provider.
type Endpoints struct {
	AuthURL  string
	TokenURL string
	UserURL  string
}

// strategy maps the configured endpoints and the user JSON to the OAuth2 flow.
type strategy struct {
	endpoints Endpoints
}

// userResponse is the user JSON of the provider.
type userResponse struct {
	Name  string `json:"name"`
	Login string `json:"userid"`
}

// New returns the generic OAuth2 sign in.
func New(db *gorm.DB, e Endpoints, s oauthflow.Settings) (*oauthflow.Flow, error) {
	return oauthflow.New(db, strategy{endpoints: e}, s)
}

func (strategy) Name() string { return ProviderName }

func (s strategy) Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{AuthURL: s.endpoints.AuthURL, TokenURL: s.endpoints.TokenURL}
}

func (strategy) SuccessRedirect() string { return "" }

// Profile loads the user from the configured user endpoint.
func (s strategy) Profile(ctx context.Context, client *http.Client) (oauthflow.Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoints.UserURL, nil)
	if err != nil {
		return oauthflow.Profile{}, fmt.Errorf("user request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return oauthflow.Profile{}, fmt.Errorf("get user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var user userResponse
	// an unreadable body leaves the user empty, as before
	_ = json.NewDecoder(resp.Body).Decode(&user)
	return oauthflow.Profile{ProviderUserID: user.Login, Login: user.Login, Name: user.Name}, nil
}
