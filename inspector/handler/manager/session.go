package manager

import (
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

type Session interface {
	GetSessionID() string
	GetLastAccessTime() time.Time
	GetAppParameter() *app.Parameter
	GetDevice() device.Device
}

type SessionManager interface {
	GetSessions() []Session
	GetSessionDetails(sessionID string) Session
}
