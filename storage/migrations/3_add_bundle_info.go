package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addBundleInfo adds the info column to app binaries.
func addBundleInfo() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddBundleInfo", Migrate: migrateAddBundleInfo}
}

func migrateAddBundleInfo(g *gorm.DB) error {
	type AppBinary struct {
		Info string
	}
	return addColumnIfMissing(g, &AppBinary{}, "Info")
}
