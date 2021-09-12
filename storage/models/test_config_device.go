package models

import (
	"gorm.io/gorm"
)

type TestConfigDevice struct {
	gorm.Model
	TestConfigID uint
	DeviceID     uint
}
