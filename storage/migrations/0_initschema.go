package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func InitSchema(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&models.Company{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.Project{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.AccessToken{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.User{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.UserAuth{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.App{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.AppBinary{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.Device{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.ConnectionParameter{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.DeviceParameter{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.DeviceLog{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.Test{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestConfig{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestConfigDevice{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestConfigScenario{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.ScenarioStep{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestConfigUnity{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.UnityTestFunction{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestProtocol{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.ProtocolEntry{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.ProtocolPerformanceEntry{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestRun{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestRunLogEntry{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&models.TestRunDeviceStatus{}); err != nil {
		return err
	}

	return nil
}
