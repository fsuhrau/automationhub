package models

import "gorm.io/gorm"

type TestRun struct {
	gorm.Model
	TestID uint
	Parameter TestParameter
	Log TestLog
	Result TestResult
}
