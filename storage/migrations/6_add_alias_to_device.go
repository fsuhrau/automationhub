package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddAliasToDevice *gormigrate.Migration

func init() {
	AddAliasToDevice = &gormigrate.Migration{
		ID: "AddAliasToDevice",
		Migrate: func(tx *gorm.DB) error {
			type Device struct {
				gorm.Model
				Alias string
			}

			if err := tx.Migrator().AddColumn(&Device{}, "Alias"); err != nil {
				return err
			}

			return nil
		},
	}
}
