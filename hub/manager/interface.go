package manager

import (
	"context"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/sirupsen/logrus"
)

type Devices interface {
	Devices() ([]device.Device, error)
	SendAction(dev device.Device, a action.Interface) error
	// UnlockDevice(s Session) error
	// LockDevice(s Session, properties *device.Properties) (*DeviceLock, error)
	//GetDevice(id uint) (device.Device, error)
	GetDevice(id string) device.Device
	Start(getDevice device.Device) error
	Stop(getDevice device.Device) error
	AddHandler(handler device.Handler)
	Run(ctx context.Context) error
	StopObserver()
	FindDevice(params *app.Parameter) device.Device
	RegisterDevice(data device.RegisterData) (device.Device, error)
}

type Sessions interface {
	GetSession(sessionID string) (Session, error)
	CreateNewSession(logger *logrus.Entry, properties *device.Properties, parameter *app.Parameter) Session
	AddSession(s Session)
	StopSession(s Session) error
	Run(ctx context.Context)
	GetSessions() []Session
	GetSessionDetails(sessionID string) Session
}

type Storage interface {
	StoreSceneGraph(data []byte) (string, error)
	StoreImage(content []byte) (string, error)
	Close() error
}
