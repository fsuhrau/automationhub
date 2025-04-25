package models

type TestConfigUnity struct {
	Model
	TestConfigID          uint                  `json:"testConfigId"`
	UnityTestCategoryType UnityTestCategoryType `json:"testCategoryType"`
	UnityTestFunctions    []UnityTestFunction   `json:"testFunctions"`
	Categories            string                `json:"categories"`
	PlayMode              bool                  `json:"playMode"`
}
