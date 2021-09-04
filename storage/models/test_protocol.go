package models

import (
	"gorm.io/gorm"
	"time"
)

type TestProtocol struct {
	gorm.Model
	TestRunID uint
	StartedAt time.Time
	EndedAt   *time.Time
	Entries   []ProtocolEntry
}
