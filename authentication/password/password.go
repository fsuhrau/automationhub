// Package password signs users in with an email address and a password.
package password

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/usersession"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// ProviderName is stored with password users and reported as sign in method.
	ProviderName = "password"
	// MinPasswordLength is the shortest password a new account can have
	MinPasswordLength = 8
	// loginURL is the front end page with the sign in form.
	loginURL = "/login"
)

// Provider is the email and password sign in.
type Provider struct {
	db    *gorm.DB
	guard *usersession.Guard
}

var _ authentication.Provider = (*Provider)(nil)

// New returns the password sign in; the session cookies are signed with secret.
func New(db *gorm.DB, secret []byte) *Provider {
	return &Provider{
		db:    db,
		guard: usersession.New(secret, usersession.Options{Method: ProviderName, LoginURL: loginURL}),
	}
}

// Session returns the middleware that loads the cookie session.
func (p *Provider) Session(name string) gin.HandlerFunc { return p.guard.Middleware(name) }

// SessionHandler returns the guard for the authenticated routes.
func (p *Provider) SessionHandler() gin.HandlerFunc { return p.guard.RequireUser() }

// Routes registers the login, register, session and logout routes.
func (p *Provider) Routes(r *gin.Engine) {
	r.POST("/auth/login", p.login)
	r.POST("/auth/register", p.register)
	p.guard.Register(r)
}
