package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// schemaModels lists every model in creation order.
func schemaModels() []any {
	return []any{
		&models.Company{}, &models.Project{}, &models.AccessToken{}, &models.User{},
		&models.UserAuth{}, &models.App{}, &models.AppParameter{}, &models.AppBinary{},
		&models.Node{}, &models.Device{}, &models.ConnectionParameter{},
		&models.DeviceParameter{}, &models.CustomParameter{}, &models.DeviceLog{},
		&models.Test{}, &models.TestConfig{}, &models.TestConfigDevice{},
		&models.TestConfigScenario{}, &models.ScenarioStep{}, &models.TestConfigUnity{},
		&models.UnityTestFunction{}, &models.TestProtocol{}, &models.ProtocolEntry{},
		&models.ProtocolPerformanceEntry{}, &models.TestRun{}, &models.TestRunLogEntry{},
		&models.TestRunDeviceStatus{}, &models.TestRunRecording{}, &models.ProtocolKey{},
	}
}

// InitSchema creates the full schema of a new database.
func InitSchema(tx *gorm.DB) error {
	for _, model := range schemaModels() {
		if err := tx.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}
