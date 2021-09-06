package models

import "gorm.io/gorm"

type TestRun struct {
	gorm.Model
	TestID    uint
	SessionID string
	Protocols []TestProtocol
	Log       []TestRunLogEntry
}
