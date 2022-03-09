package models

import "gorm.io/gorm"

type TestConfigUnity struct {
	gorm.Model
	TestConfigID       uint
	RunAllTests        bool
	UnityTestFunctions []UnityTestFunction
	Categories         string
	PlayMode           bool
}
