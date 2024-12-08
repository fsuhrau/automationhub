package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddNodes *gormigrate.Migration

func init() {
	AddNodes = &gormigrate.Migration{
		ID: "AddNodes",
		Migrate: func(tx *gorm.DB) error {
			type Node struct {
				gorm.Model
				Identifier string
				Name       string
			}

			type Device struct {
				gorm.Model
				NodeID uint
				Node   *Node
			}

			if err := tx.AutoMigrate(&Node{}); err != nil {
				return err
			}

			if err := tx.Migrator().AddColumn(&Device{}, "NodeID"); err != nil {
				return err
			}

			return nil
		},
	}
}
