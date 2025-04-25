package models

type TestConfigScenario struct {
	Model
	TestConfigID uint           `json:"testConfigId"`
	Steps        []ScenarioStep `json:"steps"`
}
