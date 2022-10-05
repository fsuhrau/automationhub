package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddOsInfosToDevice *gormigrate.Migration

func init() {
	AddOsInfosToDevice = &gormigrate.Migration{
		ID: "AddOsInfosToDevice",
		Migrate: func(tx *gorm.DB) error {

			type Device struct {
				gorm.Model
				OSInfos string
			}
			if err := tx.Migrator().AddColumn(&Device{}, "OSInfos"); err != nil {
				return err
			}

			return nil
		},
	}
}
