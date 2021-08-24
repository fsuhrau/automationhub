package api

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
	"time"
)

type Session struct {
	logger       *logrus.Logger
	sessionID    string
	lock         *manager.DeviceLock
	lastAccess   time.Time
	appParameter *app.Parameter
	storage      manager.Storage
}

func (s *Session) GetLogger() *logrus.Entry {
	return s.logger.WithField("session", s.GetSessionID())
}

func (s *Session) GetSessionID() string {
	return s.sessionID
}

func (s *Session) SetDeviceLock(lock *manager.DeviceLock) {
	s.lock = lock
}

func (s *Session) GetDeviceLock() *manager.DeviceLock {
	return s.lock
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

func (s *Session) GetAppParameter() *app.Parameter {
	return s.appParameter
}

func (s *Session) Close() error {
	return nil
}

func (s *Session) GetStorage() manager.Storage {
	return s.storage
}

func (s *Session) HandleDisconnect() {

}