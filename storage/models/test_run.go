package models

import "gorm.io/gorm"

type TestRun struct {
	gorm.Model
	TestID       uint
	Test         *Test
	AppBinaryID  uint
	AppBinary    *AppBinary
	SessionID    string
	Parameter    string
	Protocols    []TestProtocol
	Log          []TestRunLogEntry
	DeviceStatus []TestRunDeviceStatus
}
