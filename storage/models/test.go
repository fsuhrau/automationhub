package models

import "gorm.io/gorm"

type Test struct {
	gorm.Model
	AppID      uint
	App        *App
	Name       string
	TestConfig TestConfig
	TestRuns   []TestRun
	Last       *TestRun
}
