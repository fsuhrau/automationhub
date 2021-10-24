package models

import (
	"gorm.io/gorm"
)

type ProtocolPerformanceEntry struct {
	gorm.Model
	TestProtocolID uint
	Checkpoint     string
	FPS            float32   `sql:"type:decimal(10,2);"`
	MEM            float32   `sql:"type:decimal(10,2);"`
	CPU            float32   `sql:"type:decimal(10,2);"`
	Other          string
}
