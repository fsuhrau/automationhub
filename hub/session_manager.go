package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const (
	SessionTimeout = time.Second * 60
)

var (
	UnkownSessionError = fmt.Errorf("unkown session")
)

type SessionManager struct {
	dm       *DeviceManager
	sessions map[string]*Session
	log      *logrus.Entry
	kill     bool
}

func NewSessionManager(logger *logrus.Logger, deviceManager *DeviceManager) *SessionManager {
	return &SessionManager{
		log: logger.WithFields(logrus.Fields{
			"prefix": "session",
		}),
		dm:       deviceManager,
		sessions: make(map[string]*Session),
	}
}

func (s *SessionManager) cleanupSessions() {
	for sessionID, session := range s.sessions {
		if time.Now().Sub(session.LastAccess) > SessionTimeout {
			s.log.Warningf("session %s expired", sessionID)
			s.StopSession(session)
		}
	}
}

func (s *SessionManager) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.log.Infof("Stop SessionManager")
				return
			default:
			}
			s.cleanupSessions()
			if s.kill == true {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (s *SessionManager) CreateNewSession(log *logrus.Logger, properties *device.Properties, appParameter *app.Parameter) *Session {
	u, _ := uuid.NewV4()
	sessionID := fmt.Sprintf("%s", u)
	session := &Session{
		SessionID:        sessionID,
		logger:           log.WithField("session", sessionID),
		DeviceProperties: properties,
		AppParameter:     appParameter,
		Storage:          NewSessionStorage("logs", sessionID),
	}
	s.log.Infof("create new session %s", sessionID)
	return session
}

func (s *SessionManager) AddSession(session *Session) {
	session.LastAccess = time.Now()
	s.sessions[session.SessionID] = session
}

func (s *SessionManager) StopSession(session *Session) error {
	if session != nil {
		s.log.Infof("stop session %s", session.SessionID)
		if err := session.Close(); err != nil {
			s.log.Errorf("Session can't be cleaned: %v", err)
		}
		s.dm.UnlockDevice(session)
		session.Storage.Close()
		delete(s.sessions, session.SessionID)
	}
	return nil
}

func (s *SessionManager) GetSession(sessionID string) (*Session, error) {
	if session, ok := s.sessions[sessionID]; ok {
		return session, nil
	}
	return nil, UnkownSessionError
}
