package models

import (
	"gorm.io/gorm"
	"time"
)

type TestProtocol struct {
	gorm.Model
	TestRunID   uint
	TestRun     *TestRun
	DeviceID    *uint
	Device      *Device
	TestName    string
	StartedAt   time.Time
	EndedAt     *time.Time
	Entries     []ProtocolEntry
	TestResult  TestResultState
	Performance []ProtocolPerformanceEntry
	AvgFPS      float32 `sql:"type:decimal(10,2);"`
	AvgMEM      float32 `sql:"type:decimal(10,2);"`
	AvgCPU      float32 `sql:"type:decimal(10,2);"`
	// calculated fields
	HistAvgFPS          float32        `db:"-" gorm:"-:all"`
	HistAvgMEM          float32        `db:"-" gorm:"-:all"`
	HistAvgCPU          float32        `db:"-" gorm:"-:all"`
	TestProtocolHistory []TestProtocol `db:"-" gorm:"-:all"`
}
