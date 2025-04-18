package models

import "gorm.io/gorm"

type TestType uint

const (
	TestTypeUnity TestType = iota
	TestTypeCocos
	TestTypeSerenity
	TestTypeScenario
)

type ExecutionType uint

const (
	ConcurrentExecutionType     ExecutionType = iota // each test on a different free device ( faster results )
	SimultaneouslyExecutionType                      // every test on every device ( more accuracy )
)

type UnityTestCategoryType uint

const (
	AllTest UnityTestCategoryType = iota
	AllOfCategory
	SelectedTestsOnly
)

type TestConfig struct {
	gorm.Model
	TestID        uint
	ExecutionType ExecutionType
	Type          TestType
	AllDevices    bool
	Devices       []TestConfigDevice
	Unity         *TestConfigUnity
	// Cocos 	*CocosTestConfig
	// Serenity *SerenityTestConfig
	Scenario *TestConfigScenario
}

func (t *TestConfig) GetDeviceIds() []uint {
	var deviceIds []uint
	for _, d := range t.Devices {
		deviceIds = append(deviceIds, d.DeviceID)
	}
	return deviceIds
}
