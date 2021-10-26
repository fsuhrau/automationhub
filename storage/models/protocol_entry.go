package models

import (
	"gorm.io/gorm"
	"time"
)

type ProtocolEntry struct {
	gorm.Model
	TestProtocolID uint
	Timestamp      time.Time
	Source         string
	Level          string
	Message        string
	Data           string
	Runtime        int64
}
