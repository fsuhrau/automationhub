package models

type TestRunDeviceStatus struct {
	Model
	TestRunID   uint    `json:"testRunId"`
	DeviceID    uint    `json:"deviceId"`
	Device      *Device `json:"device"`
	StartupTime uint    `json:"startupTime"`

	// calculated fields
	HistAvgStartupTime uint `db:"-" gorm:"-:all" json:"histAvgStartupTime"`
}
