package teststore

import (
	"context"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// CreateTest stores the test with its config, devices, Unity settings and
// test functions in one transaction; nothing is stored when a step fails.
// The returned test has no config loaded.
func (s *Store) CreateTest(ctx context.Context, test models.Test) (models.Test, error) {
	cfg := test.TestConfig
	test.TestConfig = models.TestConfig{}
	err := s.session(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&test).Error; err != nil {
			return err
		}
		return createConfig(tx, test.ID, cfg)
	})
	return test, storeerr.Map(err, "test")
}

func createConfig(tx *gorm.DB, testID uint, cfg models.TestConfig) error {
	devices, unity := cfg.Devices, cfg.Unity
	cfg.TestID, cfg.Devices, cfg.Unity = testID, nil, nil
	if err := tx.Create(&cfg).Error; err != nil {
		return err
	}
	if err := createDevices(tx, cfg.ID, devices); err != nil {
		return err
	}
	return createUnity(tx, cfg.ID, unity)
}

func createDevices(tx *gorm.DB, configID uint, devices []models.TestConfigDevice) error {
	for _, d := range devices {
		if err := tx.Create(&models.TestConfigDevice{TestConfigID: configID, DeviceID: d.DeviceID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func createUnity(tx *gorm.DB, configID uint, unity *models.TestConfigUnity) error {
	if unity == nil {
		return nil
	}
	functions := unity.UnityTestFunctions
	row := models.TestConfigUnity{TestConfigID: configID, UnityTestCategoryType: unity.UnityTestCategoryType, Categories: unity.Categories}
	if err := tx.Create(&row).Error; err != nil {
		return err
	}
	return createFunctions(tx, row.ID, functions)
}

func createFunctions(tx *gorm.DB, unityID uint, functions []models.UnityTestFunction) error {
	for _, f := range functions {
		f.TestConfigUnityID = unityID
		if err := tx.Create(&f).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateTest saves the desired name and config over the loaded test in one
// transaction: the devices and, when the test has Unity settings, the Unity
// test functions are replaced. Nothing is saved when a step fails.
func (s *Store) UpdateTest(ctx context.Context, current *models.Test, desired models.Test) error {
	err := s.session(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Test{}).Where("id = ?", current.ID).Update("name", desired.Name).Error; err != nil {
			return err
		}
		return updateConfig(tx, current.TestConfig, desired.TestConfig)
	})
	return storeerr.Map(err, "test")
}

func updateConfig(tx *gorm.DB, current, desired models.TestConfig) error {
	if err := tx.Model(&models.TestConfig{}).Where("id = ?", current.ID).
		Updates(map[string]interface{}{"execution_type": desired.ExecutionType, "all_devices": desired.AllDevices, "record_session": desired.RecordSession}).Error; err != nil {
		return err
	}
	if err := tx.Where("test_config_id = ?", current.ID).Delete(&models.TestConfigDevice{}).Error; err != nil {
		return err
	}
	if err := createDevices(tx, current.ID, desired.Devices); err != nil {
		return err
	}
	if desired.Unity == nil || current.Unity == nil {
		return nil
	}
	return updateUnity(tx, current.Unity.ID, *desired.Unity)
}

func updateUnity(tx *gorm.DB, unityID uint, desired models.TestConfigUnity) error {
	if err := tx.Model(&models.TestConfigUnity{}).Where("id = ?", unityID).
		Updates(map[string]interface{}{"unity_test_category_type": desired.UnityTestCategoryType, "categories": desired.Categories}).Error; err != nil {
		return err
	}
	if err := tx.Where("test_config_unity_id = ?", unityID).Delete(&models.UnityTestFunction{}).Error; err != nil {
		return err
	}
	return createFunctions(tx, unityID, desired.UnityTestFunctions)
}
