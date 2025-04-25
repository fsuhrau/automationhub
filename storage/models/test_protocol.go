package models

import (
	"time"
)

type TestProtocol struct {
	Model
	TestRunID            uint                       `json:"testRunId,omitempty"`
	ParentTestProtocolID *uint                      `sql:"default:NULL" json:"parentTestProtocolId,omitempty"`
	TestRun              *TestRun                   `json:"testRun,omitempty"`
	DeviceID             *uint                      `json:"deviceId,omitempty"`
	Device               *Device                    `json:"device,omitempty"`
	TestName             string                     `json:"testName"`
	StartedAt            time.Time                  `json:"startedAt,omitempty"`
	EndedAt              *time.Time                 `json:"endedAt,omitempty"`
	Entries              []ProtocolEntry            `json:"entries,omitempty"`
	TestResult           TestResultState            `json:"testResult,omitempty"`
	Performance          []ProtocolPerformanceEntry `json:"performance,omitempty"`
	AvgFPS               float64                    `sql:"type:decimal(10,2);" json:"avgFps,omitempty"`
	AvgMEM               float64                    `sql:"type:decimal(10,2);" json:"avgMem,omitempty"`
	AvgCPU               float64                    `sql:"type:decimal(10,2);" json:"avgCpu,omitempty"`
	AvgVertexCount       float64                    `sql:"type:decimal(10,2);" json:"avgVertexCount,omitempty"`
	AvgTriangles         float64                    `sql:"type:decimal(10,2);" json:"avgTriangles,omitempty"`
	// calculated fields
	HistAvgFPS          float64        `db:"-" gorm:"-:all" json:"histAvgFps,omitempty"`
	HistAvgMEM          float64        `db:"-" gorm:"-:all" json:"histAvgMem,omitempty"`
	HistAvgCPU          float64        `db:"-" gorm:"-:all" json:"histAvgCpu,omitempty"`
	HistAvgVertexCount  float64        `db:"-" gorm:"-:all" json:"histAvgVertexCount,omitempty"`
	HistAvgTriangles    float64        `db:"-" gorm:"-:all" json:"histAvgTriangles,omitempty"`
	TestProtocolHistory []TestProtocol `db:"-" gorm:"-:all" json:"testProtocolHistory,omitempty"`
}
