package models

import "gorm.io/gorm"

type StepType uint

const (
	StepTypeInstallApp StepType = iota
	StepTypeUninstallApp
	StepTypeStartApp
	StepTypeStopApp
	StepTypeExecuteTest
	StepTypeCheckpoint
)

type ScenarioStep struct {
	gorm.Model
	TestConfigScenarioID uint
	StepType             StepType
	InstallIdentifier    string
	Checkpoint           string
	AppIdentifier        string
	TestName             string
}
