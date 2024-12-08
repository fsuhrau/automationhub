package models

import (
	"gorm.io/gorm"
)

type ProtocolPerformanceEntry struct {
	gorm.Model
	TestProtocolID uint
	Checkpoint     string
	FPS            float64 `sql:"type:decimal(10,2);"`
	MEM            float64 `sql:"type:decimal(10,2);"`
	CPU            float64 `sql:"type:decimal(10,2);"`
	Other          string
	Runtime        float64
}
