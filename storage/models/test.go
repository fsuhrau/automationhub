package models

import "gorm.io/gorm"

type Test struct {
	gorm.Model
	CompanyID  uint
	Name       string
	TestConfig TestConfig
	TestRuns   []TestRun
	Last       *TestRun
}
