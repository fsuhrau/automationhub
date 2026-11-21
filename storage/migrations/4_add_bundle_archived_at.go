package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addBundleArchivedAt adds the archived_at column to app binaries.
func addBundleArchivedAt() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddBundleArchivedAt", Migrate: migrateAddBundleArchivedAt}
}

func migrateAddBundleArchivedAt(g *gorm.DB) error {
	type AppBinary struct {
		ArchivedAt *time.Time
	}
	return addColumnIfMissing(g, &AppBinary{}, "ArchivedAt")
}
