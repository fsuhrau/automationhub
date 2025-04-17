package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddParentProtocolIdToTestProtocol *gormigrate.Migration

func init() {
	AddParentProtocolIdToTestProtocol = &gormigrate.Migration{
		ID: "AddParentProtocolIdToTestProtocol",
		Migrate: func(tx *gorm.DB) error {
			type TestProtocol struct {
				gorm.Model
				ParentTestProtocolID *uint `sql:"default:NULL"`
			}

			if err := tx.Migrator().AddColumn(&TestProtocol{}, "ParentTestProtocolID"); err != nil {
				return err
			}

			return nil
		},
	}
}
