package models

import (
	"gorm.io/gorm"
	"time"
)

type TestProtocol struct {
	gorm.Model
	TestRunID            uint  `json:",omitempty"`
	ParentTestProtocolID *uint `sql:"default:NULL" json:",omitempty"`
	TestRun              *TestRun
	DeviceID             *uint `json:",omitempty"`
	Device               *Device
	TestName             string
	StartedAt            time.Time
	EndedAt              *time.Time
	Entries              []ProtocolEntry
	TestResult           TestResultState
	Performance          []ProtocolPerformanceEntry
	AvgFPS               float64 `sql:"type:decimal(10,2);" json:",omitempty"`
	AvgMEM               float64 `sql:"type:decimal(10,2);" json:",omitempty"`
	AvgCPU               float64 `sql:"type:decimal(10,2);" json:",omitempty"`
	// calculated fields
	HistAvgFPS          float64        `db:"-" gorm:"-:all" json:",omitempty"`
	HistAvgMEM          float64        `db:"-" gorm:"-:all" json:",omitempty"`
	HistAvgCPU          float64        `db:"-" gorm:"-:all" json:",omitempty"`
	TestProtocolHistory []TestProtocol `db:"-" gorm:"-:all" json:",omitempty"`
}
