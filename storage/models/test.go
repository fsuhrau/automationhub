package models

type Test struct {
	Model
	AppID      uint       `json:"appId"`
	App        *App       `json:"app"`
	Name       string     `json:"name"`
	TestConfig TestConfig `json:"testConfig"`
	TestRuns   []TestRun  `json:"testRuns"`
	Last       *TestRun   `json:"last"`
}
