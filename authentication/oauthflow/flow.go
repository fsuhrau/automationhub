package oauthflow

import (
	"fmt"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/usersession"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// loginPath is where the front end starts a sign in and where the provider returns.
const loginPath = "/auth/oauth2"

// Flow is a sign in provider that works with one Strategy.
type Flow struct {
	conf     *oauth2.Config
	strategy Strategy
	users    users
	guard    *usersession.Guard
}

var _ authentication.Provider = (*Flow)(nil)

// New builds the flow of a strategy from the configured settings.
func New(db *gorm.DB, strategy Strategy, s Settings) (*Flow, error) {
	creds, err := LoadCredentials(s.CredFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", strategy.Name(), err)
	}
	return &Flow{
		conf: &oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  s.RedirectURL,
			Scopes:       s.Scopes,
			Endpoint:     strategy.Endpoint(),
		},
		strategy: strategy,
		users:    users{db: db, provider: strategy.Name()},
		guard: usersession.New(s.Secret, usersession.Options{
			Method: strategy.Name(), LoginURL: loginPath, StatusIncludesURL: true,
		}),
	}, nil
}

// Session returns the middleware that loads the cookie session.
func (f *Flow) Session(name string) gin.HandlerFunc { return f.guard.Middleware(name) }

// SessionHandler returns the guard for the authenticated routes.
func (f *Flow) SessionHandler() gin.HandlerFunc { return f.guard.RequireUser() }

// Routes registers the session, logout and OAuth2 routes.
func (f *Flow) Routes(r *gin.Engine) {
	f.guard.Register(r)
	r.Any(loginPath, f.login)
}
