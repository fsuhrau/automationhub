package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddUnityTestCategory *gormigrate.Migration

func init() {
	AddUnityTestCategory = &gormigrate.Migration{
		ID: "AddUnityTestCategory",
		Migrate: func(tx *gorm.DB) error {
			type TestConfigUnity struct {
				gorm.Model
				UnityTestCategoryType models.UnityTestCategoryType
			}

			if err := tx.Migrator().AddColumn(&TestConfigUnity{}, "UnityTestCategoryType"); err != nil {
				return err
			}

			return nil
		},
	}
}
