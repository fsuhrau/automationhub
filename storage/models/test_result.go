package models

import "gorm.io/gorm"

type TestResultState uint

type TestResult struct {
	gorm.Model
	TestRunID uint
	Status TestResultState
}
