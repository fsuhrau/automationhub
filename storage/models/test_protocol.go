package models

import (
	"gorm.io/gorm"
	"time"
)

type TestProtocol struct {
	gorm.Model
	TestRunID  uint
	DeviceID   *uint
	Device     *Device
	TestName   string
	StartedAt  time.Time
	EndedAt    *time.Time
	Entries    []ProtocolEntry
	TestResult TestResultState
}
