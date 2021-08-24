package selenium

import (
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
	"time"
)

type Session struct {
	logger       *logrus.Logger
	XmlDocument  *xmlquery.Node
	timeout      time.Duration
	lock         hub.DeviceLock
	AppParameter *app.Parameter
}

func (s *Session) GetLogger() *logrus.Entry {
	return s.logger.WithField("session", s.GetSessionID())
}

func (s *Session) GetActionTimeout() time.Time {
	return time.Now().Add(s.timeout)
}

func (s *Session) GetSessionID() string {
	return ""
}

func (s *Session) SetActionTimeout(duration time.Duration) {
	s.timeout = duration
}

func (s *Session) GetDevice() device.Device {
	return s.lock.Device
}

func (s *Session) GetStorage() manager.Storage {
	return nil
}

func (s *Session) WaitForConnection() {

}
