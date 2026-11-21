package wiring

import (
	"fmt"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/password"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub"
	"gorm.io/gorm"
)

// buildAuth sets up the first configured authentication provider (OAuth2,
// GitHub, password). Without one the hub stays open.
func buildAuth(db *gorm.DB, auth config.Auth) (hub.Auth, error) {
	provider, err := authProvider(db, auth)
	if err != nil {
		return hub.Auth{}, fmt.Errorf("set up authentication: %w", err)
	}
	if provider == nil {
		return hub.Auth{}, nil
	}
	return hub.Auth{
		Session: provider.Session("session"),
		Guard:   token.SessionHandler(db, provider.SessionHandler()),
		Routes:  provider.Routes,
	}, nil
}

// authProvider builds the provider of the first configured method, nil for none.
func authProvider(db *gorm.DB, auth config.Auth) (authentication.Provider, error) {
	switch {
	case auth.OAuth2 != nil:
		o := auth.OAuth2
		endpoints := oauth2.Endpoints{AuthURL: o.AuthUrl, TokenURL: o.TokenUrl, UserURL: o.UserUrl}
		return oauth2.New(db, endpoints, oauthflow.Settings{RedirectURL: o.RedirectUrl, CredFile: o.Credentials, Scopes: o.Scopes, Secret: []byte(o.Secret)})
	case auth.Github != nil:
		g := auth.Github
		return github.New(db, oauthflow.Settings{RedirectURL: g.RedirectUrl, CredFile: g.Credentials, Scopes: g.Scopes, Secret: []byte(g.Secret)})
	case auth.Password != nil:
		return password.New(db, []byte(auth.Password.Secret)), nil
	}
	return nil, nil
}
