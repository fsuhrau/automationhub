package hub

import (
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

var (
	DEVICE_CONNECTION_TIMEOUT = 30 * time.Second
	CONNECTION_TIMEOUT        = fmt.Errorf("connection timeout")
	ActionTimeout             = 20 * time.Second
)

type Recorder struct {
	Storage *SessionStorage
	Device  device.Device
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
	DeviceProperties *device.Properties
	AppParameter     *app.Parameter
	XmlDocument      *xmlquery.Node
	Storage          *SessionStorage
	actionTimeout    time.Duration
	endSession       bool
}

func (s *Session) GetSessionID() string {
	return s.SessionID
}

func (s *Session) GetLastAccessTime() time.Time {
	return s.LastAccess
}

func (s *Session) GetAppParameter() *app.Parameter {
	return s.AppParameter
}

func (s *Session) GetDevice() device.Device {
	if s.Lock != nil {
		return s.Lock.Device
	}
	return nil
}

func (s *Session) WaitForConnection() error {

	timeout := time.Now().Add(s.Lock.Device.ConnectionTimeout())

	for !s.Lock.Device.IsAppConnected() {
		if time.Now().After(timeout) {
			return CONNECTION_TIMEOUT
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *Session) SetActionTimeout(timeout time.Duration) {
	s.actionTimeout = timeout
}

func (s *Session) GetActionTimeout() time.Time {
	var timeout time.Duration
	if s.actionTimeout > 0 {
		timeout = s.actionTimeout
	} else {
		timeout = ActionTimeout
	}
	return time.Now().Add(timeout)
}

func (s *Session) Close() error {
	s.endSession = true
	if s.Recorder != nil {
		if err := s.Recorder.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) DeviceDisconnected() {
	if s.endSession != true {
		s.logger.Errorln("Device Crashed!")
	}
}
