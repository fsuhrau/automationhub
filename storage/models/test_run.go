package models

import "gorm.io/gorm"

type TestRun struct {
	gorm.Model
	TestID    uint
	AppID     uint
	App       App
	SessionID string
	Protocols []TestProtocol
	Log       []TestRunLogEntry
}
