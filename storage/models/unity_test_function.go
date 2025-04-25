package models

type UnityTestFunction struct {
	Model
	TestConfigUnityID uint   `json:"testConfigUnityId"`
	Assembly          string `json:"assembly"`
	Class             string `json:"class"`
	Method            string `json:"method"`
}
