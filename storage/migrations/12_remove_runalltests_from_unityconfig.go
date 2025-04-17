package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var RemoveRunAllTestsFromUnityConfig *gormigrate.Migration

func init() {
	RemoveRunAllTestsFromUnityConfig = &gormigrate.Migration{
		ID: "RemoveRunAllTestsFromUnityConfig",
		Migrate: func(tx *gorm.DB) error {
			type TestConfigUnity struct {
				gorm.Model
				RunAllTests bool
			}

			if err := tx.Migrator().DropColumn(&TestConfigUnity{}, "RunAllTests"); err != nil {
				return err
			}
			return nil
		},
	}
}
