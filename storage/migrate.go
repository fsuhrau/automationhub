package storage

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.App{}); err != nil {
		fmt.Printf("App: %v\n", err)
	}
	if err := db.AutoMigrate(&models.AppFunction{}); err != nil {
		fmt.Printf("AppFunction: %v\n", err)
	}
	if err := db.AutoMigrate(&models.Company{}); err != nil {
		fmt.Printf("Company: %v\n", err)
	}
	if err := db.AutoMigrate(&models.Device{}); err != nil {
		fmt.Printf("Device: %v\n", err)
	}
	if err := db.AutoMigrate(&models.Test{}); err != nil {
		fmt.Printf("Test: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestConfig{}); err != nil {
		fmt.Printf("TestConfig: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestRun{}); err != nil {
		fmt.Printf("TestRun: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestRunLogEntry{}); err != nil {
		fmt.Printf("TestRunLogEntry: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestProtocol{}); err != nil {
		fmt.Printf("TestProtocol: %v\n", err)
	}
	if err := db.AutoMigrate(&models.ProtocolEntry{}); err != nil {
		fmt.Printf("ProtocolEntry: %v\n", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		fmt.Printf("User: %v\n", err)
	}
	if err := db.AutoMigrate(&models.UserAuth{}); err != nil {
		fmt.Printf("UserAuth: %v\n", err)
	}
	if err := db.AutoMigrate(&models.DeviceLog{}); err != nil {
		fmt.Printf("DeviceLog: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestConfigUnity{}); err != nil {
		fmt.Printf("TestConfigUnity: %v\n", err)
	}
	if err := db.AutoMigrate(&models.UnityTestFunction{}); err != nil {
		fmt.Printf("UnityTestFunction: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestConfigDevice{}); err != nil {
		fmt.Printf("TestConfigDevice: %v\n", err)
	}
	if err := db.AutoMigrate(&models.ProtocolPerformanceEntry{}); err != nil {
		fmt.Printf("ProtocolPerformanceEntry: %v\n", err)
	}
	if err := db.AutoMigrate(&models.ScenarioStep{}); err != nil {
		fmt.Printf("ScenarioStep: %v\n", err)
	}
	if err := db.AutoMigrate(&models.TestConfigScenario{}); err != nil {
		fmt.Printf("TestConfigScenario: %v\n", err)
	}
	if err := db.AutoMigrate(&models.DeviceParameter{}); err != nil {
		fmt.Printf("DeviceParameter: %v\n", err)
	}
	if err := db.AutoMigrate(&models.ConnectionParameter{}); err != nil {
		fmt.Printf("ConnectionParameter: %v\n", err)
	}
	return nil
}