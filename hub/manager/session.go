package manager

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
	"time"
)

type Session interface {
	GetSessionID() string
	SetDeviceLock(lock *DeviceLock)
	GetDeviceLock() *DeviceLock
	GetDevice() device.Device
	WaitForConnection() error
	GetLastAccess() time.Time
	SetLastAccess(now time.Time)
	GetAppParameter() *app.Parameter
	Close() error
	GetStorage() Storage
	GetLogger() *logrus.Entry
	HandleDisconnect()
}
