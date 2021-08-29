package models

import (
	"gorm.io/gorm"
	"time"
)

type TestLog struct {
	gorm.Model
	TestRunID uint
	StartedAt time.Time
	EndedAt   *time.Time
	Entries []LogEntry
}
