package models

import (
	"gorm.io/gorm"
	"time"
)

type LogEntry struct {
	gorm.Model
	TestLogID uint
	Timestamp time.Time
	Source    string
	Info      string
}
