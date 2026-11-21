// Package flowtest runs the sign in scenarios every OAuth2 provider must pass
// against a fake identity provider.
package flowtest

import (
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"gorm.io/gorm"
)

// Case describes one provider under test.
type Case struct {
	// Name is the sign in method the provider reports.
	Name string
	// UserJSON is what the fake user endpoint answers; it maps to the user "octo" / "Octo Cat".
	UserJSON string
	// CallbackStatus is the status of a successful callback.
	CallbackStatus int
	// FailsOnBrokenUser tells whether an unreadable user answer refuses the login.
	FailsOnBrokenUser bool
	// Build creates the provider against the fake identity provider at idpURL.
	Build func(db *gorm.DB, idpURL string, s oauthflow.Settings) (*oauthflow.Flow, error)
}

// Run runs every scenario against the provider of c.
func Run(t *testing.T, c Case) {
	scenarios := map[string]func(*testing.T, Case){
		"login flow":                    loginFlow,
		"refuses wrong or reused state": refusesWrongOrReusedState,
		"provider failures":             providerFailures,
		"signed in user":                signedInUserPassesLoginRoute,
		"concurrent logins":             concurrentLoginsDoNotInterfere,
		"session answers":               sessionAnswersNameTheMethod,
	}
	for name, run := range scenarios {
		t.Run(name, func(t *testing.T) { run(t, c) })
	}
}
