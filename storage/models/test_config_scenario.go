package models

import "gorm.io/gorm"

type TestConfigScenario struct {
	gorm.Model
	TestConfigID uint
	Steps        []ScenarioStep
}
