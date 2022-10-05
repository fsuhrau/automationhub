package models

import "gorm.io/gorm"

type TestRunDeviceStatus struct {
	gorm.Model
	TestRunID   uint
	DeviceID    uint
	Device      *Device
	StartupTime uint

	// calculated fields
	HistAvgStartupTime uint `db:"-" gorm:"-:all"`
}
