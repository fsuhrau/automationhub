package models

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
	Model
	TestID        uint               `json:"testId"`
	Test          *Test              `json:"test"`
	ExecutionType ExecutionType      `json:"executionType"`
	Type          TestType           `json:"type"`
	AllDevices    bool               `json:"allDevices"`
	Devices       []TestConfigDevice `json:"devices"`
	Unity         *TestConfigUnity   `json:"unity"`
	// Cocos 	*CocosTestConfig
	// Serenity *SerenityTestConfig
	Scenario *TestConfigScenario `json:"scenario"`
}

func (t *TestConfig) GetDeviceIds() []uint {
	var deviceIds []uint
	for _, d := range t.Devices {
		deviceIds = append(deviceIds, d.DeviceID)
	}
	return deviceIds
}
