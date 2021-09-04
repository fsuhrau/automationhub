package models

import "gorm.io/gorm"

type TestRun struct {
	gorm.Model
	TestID    uint
	SessionID string
	Parameter TestParameter
	Log       TestProtocol
	Result    TestResult
}
