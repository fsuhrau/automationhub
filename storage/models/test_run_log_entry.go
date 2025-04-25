package models

type TestRunLogEntry struct {
	Model
	TestRunID uint   `json:"testRunId"`
	Level     string `json:"level"`
	Log       string `json:"log"`
}
