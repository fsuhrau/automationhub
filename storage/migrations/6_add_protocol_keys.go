package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addProtocolKeys creates the table of protocol keys.
func addProtocolKeys() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddProtocolKeys", Migrate: migrateAddProtocolKeys}
}

func migrateAddProtocolKeys(g *gorm.DB) error {
	return g.AutoMigrate(&models.ProtocolKey{})
}
