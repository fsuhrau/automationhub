package selenium

import (
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
	"time"
)

type Session struct {
	logger       *logrus.Logger
	XmlDocument  *xmlquery.Node
	timeout      time.Duration
	lock         *manager.DeviceLock
	AppParameter *app.Parameter
	storage      manager.Storage
	lastAccess   time.Time
}

func (s *Session) GetLogger() *logrus.Entry {
	return s.logger.WithField("session", s.GetSessionID())
}

func (s *Session) GetSessionID() string {
	return ""
}

func (s *Session) SetDeviceLock(lock *manager.DeviceLock) {
	s.lock = lock
}

func (s *Session) GetDeviceLock() *manager.DeviceLock {
	return s.lock
}

func (s *Session) GetActionTimeout() time.Time {
	return time.Now().Add(s.timeout)
}

func (s *Session) SetActionTimeout(duration time.Duration) {
	s.timeout = duration
}

func (s *Session) GetDevice() device.Device {
	return s.lock.Device
}

func (s *Session) WaitForConnection() error {
	return nil
}

func (s *Session) GetLastAccess() time.Time {
	return s.lastAccess
}

func (s *Session) SetLastAccess(now time.Time) {
	s.lastAccess = now
}

func (s *Session) Close() error {
	return nil
}

func (s *Session) GetAppParameter() *app.Parameter {
	return s.AppParameter
}

func (s *Session) GetStorage() manager.Storage {
	return s.storage
}

func (s *Session) HandleDisconnect() {

}