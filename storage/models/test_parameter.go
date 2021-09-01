package models

import "gorm.io/gorm"

type TestParameter struct {
	gorm.Model
	TestRunID uint
	AppID     uint
	DeviceID  uint
}
