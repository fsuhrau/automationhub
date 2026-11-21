package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// Timeout is how long a session may stay unused before it is stopped.
	Timeout = 1 * time.Minute

	cleanupInterval = 50 * time.Millisecond
	logsDir         = "logs"
)

// ErrUnknown is returned for a session id that is not registered.
var ErrUnknown = errors.New("unknown session")

// Manager keeps the running sessions and stops the expired ones.
type Manager struct {
	log   *logrus.Entry
	clock Clock
	deps  Deps

	mu       sync.Mutex
	sessions map[string]manager.Session
}

// NewManager creates a manager that creates session storage with deps.
func NewManager(logger *logrus.Logger, deps Deps) *Manager {
	return &Manager{
		log:      logger.WithFields(logrus.Fields{"prefix": "session"}),
		clock:    deps.Clock,
		deps:     deps,
		sessions: make(map[string]manager.Session),
	}
}

// Run stops expired sessions in the background until ctx is done.
func (m *Manager) Run(ctx context.Context) {
	go m.cleanupLoop(ctx)
}

func (m *Manager) cleanupLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			m.log.Infof("Stop SessionManager")
			return
		}
		m.cleanupSessions()
		select {
		case <-ctx.Done():
		case <-m.clock.After(cleanupInterval):
		}
	}
}

func (m *Manager) cleanupSessions() {
	for _, s := range m.expiredSessions() {
		m.log.Warningf("session %s expired", s.GetSessionID())
		_ = m.StopSession(s)
	}
}

func (m *Manager) expiredSessions() []manager.Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	var expired []manager.Session
	for _, s := range m.sessions {
		if m.clock.Now().Sub(s.GetLastAccess()) > Timeout {
			expired = append(expired, s)
		}
	}
	return expired
}

// CreateNewSession creates a session with its own storage; it is not
// registered until AddSession.
func (m *Manager) CreateNewSession(log *logrus.Entry, properties *device.Properties, appParameter *app.Parameter) manager.Session {
	u, _ := uuid.NewV4()
	sessionID := u.String()
	m.log.Infof("create new session %s", sessionID)
	return &Session{
		SessionID:        sessionID,
		logger:           log.WithField("session", sessionID),
		clock:            m.clock,
		DeviceProperties: properties,
		AppParameter:     appParameter,
		Storage:          NewStorage(m.deps, logsDir, sessionID),
	}
}

// AddSession registers the session and marks it as just used.
func (m *Manager) AddSession(s manager.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.SetLastAccess(m.clock.Now())
	m.sessions[s.GetSessionID()] = s
}

// StopSession closes the session and its storage and forgets it.
func (m *Manager) StopSession(s manager.Session) error {
	if s == nil {
		return nil
	}
	m.log.Infof("stop session %s", s.GetSessionID())
	if err := s.Close(); err != nil {
		m.log.Errorf("Session can't be cleaned: %v", err)
	}
	if storage := s.GetStorage(); storage != nil {
		_ = storage.Close()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, s.GetSessionID())
	return nil
}

// GetSession returns the registered session or ErrUnknown.
func (m *Manager) GetSession(sessionID string) (manager.Session, error) {
	if s := m.GetSessionDetails(sessionID); s != nil {
		return s, nil
	}
	return nil, ErrUnknown
}

// GetSessions returns all registered sessions.
func (m *Manager) GetSessions() []manager.Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	var sessions []manager.Session
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetSessionDetails returns the registered session or nil.
func (m *Manager) GetSessionDetails(sessionID string) manager.Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[sessionID]
}

var _ manager.Sessions = (*Manager)(nil)
