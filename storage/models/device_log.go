package models

import (
	"github.com/fsuhrau/automationhub/device"
	"gorm.io/gorm"
	"time"
)

type DeviceLog struct {
	gorm.Model
	DeviceID uint
	Time time.Time
	Status device.State
	Payload string
}
