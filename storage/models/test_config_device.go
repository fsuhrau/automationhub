package models

type TestConfigDevice struct {
	Model
	TestConfigID uint   `json:"testConfigId"`
	DeviceID     uint   `json:"deviceId"`
	Device       Device `json:"device"`
}
