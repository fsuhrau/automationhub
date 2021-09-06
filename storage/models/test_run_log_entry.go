package models

import "gorm.io/gorm"

type TestRunLogEntry struct {
	gorm.Model
	TestRunID uint
	Level     string
	Log       string
}
