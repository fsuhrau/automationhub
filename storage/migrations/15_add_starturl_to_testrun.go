package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddStartURLToTestRun *gormigrate.Migration

func init() {
	AddStartURLToTestRun = &gormigrate.Migration{
		ID: "AddStartURLToTestRun",
		Migrate: func(tx *gorm.DB) error {
			type TestRun struct {
				gorm.Model
				StartURL string
			}

			if err := tx.Migrator().AddColumn(&TestRun{}, "StartURL"); err != nil {
				return err
			}
			return nil
		},
	}
}
