package storage

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func migrate(db *gorm.DB) error {
	db.AutoMigrate(&models.App{})
	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.Device{})
	db.AutoMigrate(&models.Test{})
	db.AutoMigrate(&models.TestConfig{})
	db.AutoMigrate(&models.TestRun{})
	db.AutoMigrate(&models.TestParameter{})
	db.AutoMigrate(&models.TestResult{})
	db.AutoMigrate(&models.TestProtocol{})
	db.AutoMigrate(&models.ProtocolEntry{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.UserAuth{})
	db.AutoMigrate(&models.DeviceLog{})
	db.AutoMigrate(&models.TestConfigUnity{})
	db.AutoMigrate(&models.UnityTestFunction{})
	return nil
}