package hub

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/devices"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

var (
	DEVICE_CONNECTION_TIMEOUT = 30 * time.Second
	CONNECTION_TIMEOUT        = fmt.Errorf("connection timeout")
)

type Recorder struct {
	Storage *SessionStorage
	Device  devices.Device
}

func (r *Recorder) Start() error {
	return r.Device.StartRecording(r.Storage.GetPath())
}

func (r *Recorder) Stop() error {
	return r.Device.StopRecording()
}

type Session struct {
	SessionID        string
	LastAccess       time.Time
	Lock             *DeviceLock
	logger           *logrus.Entry
	Recorder         *Recorder
	DeviceProperties *devices.Properties
	AppParameter     *app.Parameter
	XmlDocument      *xmlquery.Node
	Storage          *SessionStorage
}

func createNewSession(log *logrus.Logger, properties *devices.Properties, appParameter *app.Parameter) *Session {
	u, _ := uuid.NewV4()
	sessionID := fmt.Sprintf("%s", u)
	session := &Session{
		SessionID:        sessionID,
		logger:           log.WithField("session", sessionID),
		DeviceProperties: properties,
		AppParameter:     appParameter,
		Storage:          NewSessionStorage("logs", sessionID),
	}
	return session
}

func SessionMiddleware(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionID")
		if session, ok := s.sessions[sessionID]; ok {
			c.Set("session", session)
			c.Next()
		}
	}
}

func HandleWithSession(f func(*Session, *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		s, _ := c.Get("session")
		if s == nil {
			return
		}
		session := s.(*Session)
		session.LastAccess = time.Now()
		f(session, c)
	}
}

func (s *Session) WaitForConnection() error {

	// send connect command to client
	//deviceIP := s.Lock.Device.DeviceIP().String()
	//address := fmt.Sprintf("%s:6789", deviceIP)
	timeout := time.Now().Add(s.Lock.Device.ConnectionTimeout())

	for !s.Lock.Device.IsAppConnected() {
		if time.Now().After(timeout) {
			return CONNECTION_TIMEOUT
		}

		// sendOut := false
		// for !sendOut {
		// 	if time.Now().After(timeout) {
		// 		return fmt.Errorf("Connection timeout")
		// 	}

		// 	cmd := devices.NewCommand("nc", "-z", deviceIP, "6789")
		// 	_, err := cmd.Output()
		// 	if err != nil {
		// 		time.Sleep(50 * time.Millisecond)
		// 		continue
		// 	}

		// 	data, err := cctelnet.ExecCmd(address, fmt.Sprintf("INNIUM IP:%s;DEVICE_ID:%s;SESSION_ID:%s", "10.35.111.218", s.Lock.Device.DeviceID(), s.SessionID))
		// 	if err != nil {
		// 		logrus.Errorf("err: %v", err)
		// 	} else {
		// 		if !strings.Contains(data, "[Innium] success") {
		// 			continue
		// 		}
		// 		sendOut = true
		// 		break
		// 	}

		time.Sleep(1 * time.Second)
		// }
	}

	return nil
}
