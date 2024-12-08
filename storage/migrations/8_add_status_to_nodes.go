package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddStatusToNode *gormigrate.Migration

func init() {
	AddStatusToNode = &gormigrate.Migration{
		ID: "AddStatusToNode",
		Migrate: func(tx *gorm.DB) error {
			type Node struct {
				gorm.Model
				Status int32
			}

			if err := tx.Migrator().AddColumn(&Node{}, "Status"); err != nil {
				return err
			}

			return nil
		},
	}
}
