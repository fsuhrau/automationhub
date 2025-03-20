package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTargetVersionToDevice *gormigrate.Migration

func init() {
	AddTargetVersionToDevice = &gormigrate.Migration{
		ID: "AddTargetVersionToDevice",
		Migrate: func(tx *gorm.DB) error {
			type Device struct {
				gorm.Model
				TargetVersion string
			}
			if err := tx.Migrator().AddColumn(&Device{}, "TargetVersion"); err != nil {
				return err
			}

			type Android struct {
				LaunchActivity string
			}
			type Executable struct {
				Executable string
			}
			type AppBinary struct {
				gorm.Model
				Android    Android    `json:"android" db:"android" gorm:"embedded"`
				Executable Executable `json:"executable" db:"executable" gorm:"embedded"`
			}

			if err := tx.Migrator().DropColumn(&AppBinary{}, "LaunchActivity"); err != nil {
				return err
			}
			return tx.AutoMigrate(&AppBinary{})
		},
	}
}
