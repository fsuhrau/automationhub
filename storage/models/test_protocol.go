package models

import (
	"gorm.io/gorm"
	"time"
)

type TestProtocol struct {
	gorm.Model
	TestRunID  uint
	AppID      uint
	DeviceID   *uint
	Device     *Device
	StartedAt  time.Time
	EndedAt    *time.Time
	Entries    []ProtocolEntry
	TestResult TestResultState
}
