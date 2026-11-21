package session

import (
	"errors"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultActionTimeout is how long an action may take unless the session sets its own.
	DefaultActionTimeout = 20 * time.Second

	connectionPoll = time.Second
)

// ErrConnectionTimeout is returned when the app does not connect in time.
var ErrConnectionTimeout = errors.New("connection timeout")

// Recorder records the device of a session into its storage.
type Recorder struct {
	Storage *Storage
	Device  device.Device
}

// Start starts the recording.
func (r *Recorder) Start() error {
	return r.Device.StartRecording(r.Storage.GetPath())
}

// Stop stops the recording.
func (r *Recorder) Stop() error {
	return r.Device.StopRecording()
}

// Session is one test session on a device.
type Session struct {
	SessionID        string
	LastAccess       time.Time
	Lock             *manager.DeviceLock
	Recorder         *Recorder
	DeviceProperties *device.Properties
	AppParameter     *app.Parameter
	XmlDocument      *xmlquery.Node
	Storage          manager.Storage

	logger        *logrus.Entry
	clock         Clock
	actionTimeout time.Duration
	endSession    bool
}

func (s *Session) GetLogger() *logrus.Entry {
	return s.logger.WithField("session", s.GetSessionID())
}

func (s *Session) GetSessionID() string {
	return s.SessionID
}

func (s *Session) SetDeviceLock(lock *manager.DeviceLock) {
	s.Lock = lock
}

func (s *Session) GetDeviceLock() *manager.DeviceLock {
	return s.Lock
}

func (s *Session) GetDevice() device.Device {
	if s.Lock != nil {
		return s.Lock.Device
	}
	return nil
}

// WaitForConnection blocks until the app of the device is connected or the
// connection timeout of the device has passed.
func (s *Session) WaitForConnection() error {
	dev := s.Lock.Device
	deadline := s.clock.Now().Add(dev.ConnectionTimeout())
	for !dev.IsAppConnected() {
		if s.clock.Now().After(deadline) {
			return ErrConnectionTimeout
		}
		<-s.clock.After(connectionPoll)
	}
	return nil
}

func (s *Session) GetLastAccess() time.Time {
	return s.LastAccess
}

func (s *Session) SetLastAccess(t time.Time) {
	s.LastAccess = t
}

func (s *Session) GetAppParameter() *app.Parameter {
	return s.AppParameter
}

// Close ends the session and stops its recording.
func (s *Session) Close() error {
	s.endSession = true
	if s.Recorder != nil {
		return s.Recorder.Stop()
	}
	return nil
}

func (s *Session) GetStorage() manager.Storage {
	return s.Storage
}

func (s *Session) SetActionTimeout(timeout time.Duration) {
	s.actionTimeout = timeout
}

// GetActionTimeout returns the time by which an action started now has to be done.
func (s *Session) GetActionTimeout() time.Time {
	timeout := DefaultActionTimeout
	if s.actionTimeout > 0 {
		timeout = s.actionTimeout
	}
	return s.clock.Now().Add(timeout)
}

func (s *Session) DeviceDisconnected() {
	if !s.endSession {
		s.logger.Errorln("Device Crashed!")
	}
}

func (s *Session) HandleDisconnect() {}
