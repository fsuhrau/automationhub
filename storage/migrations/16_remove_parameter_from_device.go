package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var RemoveParameterFromDevice *gormigrate.Migration

func init() {
	RemoveParameterFromDevice = &gormigrate.Migration{
		ID: "RemoveParameterFromDevice",
		Migrate: func(tx *gorm.DB) error {
			type Device struct {
				gorm.Model
			}

			type CustomParameter struct {
				gorm.Model
				DeviceID uint
				Key      string
				Value    string
			}

			if err := tx.Migrator().DropColumn(&Device{}, "gpu"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&Device{}, "display_size"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&Device{}, "dpi"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&Device{}, "soc"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&Device{}, "ram"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&Device{}, "abi"); err != nil {
				return err
			}
			return tx.AutoMigrate(&CustomParameter{})
		},
	}
}
