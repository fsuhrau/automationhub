package hub

import (
	"github.com/fsuhrau/automationhub/hub/session"
	"github.com/sirupsen/logrus"
)

// SessionManager is an alias until G34 moves its users to hub/session.
type SessionManager = session.Manager

// NewSessionManager creates the session manager on the real clock and file
// system. The device manager is not used by the sessions (the unlock call
// was already commented out).
func NewSessionManager(logger *logrus.Logger, _ *DeviceManager) *SessionManager {
	return session.NewManager(logger, session.SystemDeps())
}
