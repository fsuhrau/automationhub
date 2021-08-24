package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/manager"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const (
	SessionTimeout = 1 * time.Minute
)

var (
	UnknownSessionError = fmt.Errorf("unknown session")
)

type SessionManager struct {
	dm       *DeviceManager
	sessions map[string]manager.Session
	log      *logrus.Entry
	kill     bool
}

func NewSessionManager(logger *logrus.Logger, deviceManager *DeviceManager) *SessionManager {
	return &SessionManager{
		log: logger.WithFields(logrus.Fields{
			"prefix": "session",
		}),
		dm:       deviceManager,
		sessions: make(map[string]manager.Session),
	}
}

func (s *SessionManager) cleanupSessions() {
	for sessionID, session := range s.sessions {
		if time.Now().Sub(session.GetLastAccess()) > SessionTimeout {
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

func (s *SessionManager) CreateNewSession(log *logrus.Logger, properties *device.Properties, appParameter *app.Parameter) manager.Session {
	u, _ := uuid.NewV4()
	sessionID := fmt.Sprintf("%s", u)
	sess := &Session{
		SessionID:        sessionID,
		logger:           log.WithField("session", sessionID),
		DeviceProperties: properties,
		AppParameter:     appParameter,
		Storage:          NewSessionStorage("logs", sessionID),
	}
	s.log.Infof("create new session %s", sessionID)
	return sess
}

func (s *SessionManager) AddSession(session manager.Session) {
	session.SetLastAccess(time.Now())
	s.sessions[session.GetSessionID()] = session
}

func (s *SessionManager) StopSession(session manager.Session) error {
	if session != nil {
		s.log.Infof("stop session %s", session.GetSessionID())
		if err := session.Close(); err != nil {
			s.log.Errorf("Session can't be cleaned: %v", err)
		}
		s.dm.UnlockDevice(session)
		session.GetStorage().Close()
		delete(s.sessions, session.GetSessionID())
	}
	return nil
}

func (s *SessionManager) GetSession(sessionID string) (manager.Session, error) {
	if session, ok := s.sessions[sessionID]; ok {
		return session, nil
	}
	return nil, UnknownSessionError
}

func (s *SessionManager) GetSessions() []manager.Session {
	var sessions []manager.Session
	for _, v := range s.sessions {
		sessions = append(sessions, v)
	}
	return sessions
}

func (s *SessionManager) GetSessionDetails(sessionID string) manager.Session {
	if session, ok := s.sessions[sessionID]; ok {
		return session
	}
	return nil
}
