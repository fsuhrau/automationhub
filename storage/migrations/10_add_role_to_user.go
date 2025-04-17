package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddRoleToUser *gormigrate.Migration

func init() {
	AddRoleToUser = &gormigrate.Migration{
		ID: "AddRoleToUser",
		Migrate: func(tx *gorm.DB) error {
			type User struct {
				gorm.Model
				Role string `sql:"default:NULL"`
			}

			if err := tx.Migrator().AddColumn(&User{}, "Role"); err != nil {
				return err
			}

			return nil
		},
	}
}
