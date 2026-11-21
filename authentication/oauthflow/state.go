package oauthflow

import (
	"crypto/rand"
	"crypto/subtle"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/contrib/sessions"
)

// newState returns an unguessable value for one login attempt.
func newState() string { return rand.Text() }

// storeState remembers the state of this browser's login attempt.
func storeState(session sessions.Session, state string) error {
	session.Set(authentication.StateKey, state)
	return session.Save()
}

// consumeState checks the state of the callback against the one stored for
// this browser and forgets it, so a state works only once.
func consumeState(session sessions.Session, got string) bool {
	want, _ := session.Get(authentication.StateKey).(string)
	session.Delete(authentication.StateKey)
	_ = session.Save()
	return want != "" && subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}
