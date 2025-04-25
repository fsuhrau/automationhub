package models

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
	Model
	TestConfigScenarioID uint     `json:"testConfigScenarioId"`
	StepType             StepType `json:"stepType"`
	InstallIdentifier    string   `json:"installIdentifier"`
	Checkpoint           string   `json:"checkpoint"`
	AppIdentifier        string   `json:"appIdentifier"`
	TestName             string   `json:"testName"`
}
