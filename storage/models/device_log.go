package models

import (
	"gorm.io/gorm"
	"time"
)

type DeviceLog struct {
	gorm.Model
	DeviceID uint
	Time time.Time
	Status DeviceStatus
	Payload string
}
