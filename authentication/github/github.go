// Package github signs users in with their GitHub account.
package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
	"gorm.io/gorm"
)

// ProviderName is stored with GitHub users and reported as sign in method.
const ProviderName = "github"

// Endpoints are the URLs of GitHub; they differ only for GitHub Enterprise
// and for tests.
type Endpoints struct {
	AuthURL  string
	TokenURL string
	// APIURL is the base URL of the REST API, with a trailing slash.
	APIURL string
}

// strategy maps the GitHub endpoints and users to the OAuth2 flow.
type strategy struct {
	endpoints Endpoints
}

// New returns the sign in with the public GitHub.
func New(db *gorm.DB, s oauthflow.Settings) (*oauthflow.Flow, error) {
	return NewWithEndpoints(db, Endpoints{AuthURL: oauth2gh.Endpoint.AuthURL, TokenURL: oauth2gh.Endpoint.TokenURL}, s)
}

// NewWithEndpoints returns the sign in with a GitHub at the given URLs.
func NewWithEndpoints(db *gorm.DB, e Endpoints, s oauthflow.Settings) (*oauthflow.Flow, error) {
	return oauthflow.New(db, strategy{endpoints: e}, s)
}

func (strategy) Name() string { return ProviderName }

func (s strategy) Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{AuthURL: s.endpoints.AuthURL, TokenURL: s.endpoints.TokenURL}
}

func (strategy) SuccessRedirect() string { return "/" }

// Profile loads the authenticated GitHub user.
func (s strategy) Profile(ctx context.Context, client *http.Client) (oauthflow.Profile, error) {
	api := github.NewClient(client)
	base, err := url.Parse(s.endpoints.APIURL)
	if err != nil {
		return oauthflow.Profile{}, fmt.Errorf("github api url: %w", err)
	}
	if s.endpoints.APIURL != "" {
		api.BaseURL = base
	}
	user, _, err := api.Users.Get(ctx, "")
	if err != nil {
		return oauthflow.Profile{}, fmt.Errorf("github user: %w", err)
	}
	return oauthflow.Profile{
		ProviderUserID: fmt.Sprintf("%d", user.GetID()),
		Login:          user.GetLogin(),
		Name:           user.GetName(),
		Email:          user.GetEmail(),
		URL:            user.GetURL(),
	}, nil
}
