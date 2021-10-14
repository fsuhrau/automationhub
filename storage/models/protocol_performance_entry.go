package models

import (
	"gorm.io/gorm"
)

type ProtocolPerformanceEntry struct {
	gorm.Model
	TestProtocolID uint
	Checkpoint     string
	FPS            float32
	MEM            float32
	CPU            float32
	Other          string
}
