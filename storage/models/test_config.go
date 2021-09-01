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
	ParallelExecutionType    ExecutionType = iota // each test on a different free device ( faster results )
	SynchronousExecutionType                      // every test on every device ( more accuracy )
)

type TestConfig struct {
	gorm.Model
	TestID        uint
	ExecutionType ExecutionType
	Type          TestType
	Unity         *TestConfigUnity
	// Cocos 	*CocosTestConfig
	// Serenity *SerenityTestConfig
	//Scenario *ScenarioTestConfig
}
