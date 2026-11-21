package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// All returns the upgrade migrations in the order they run. A new database
// gets InitSchema instead and counts them as done.
func All() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		addAppParameter(),
		addPersonalAccessTokens(),
		addBundleInfo(),
		addBundleArchivedAt(),
		addSessionRecording(),
		addProtocolKeys(),
	}
}

// Run brings the database to the current schema in one transaction.
func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		m := gormigrate.New(tx, gormigrate.DefaultOptions, All())
		m.InitSchema(InitSchema)
		return m.Migrate()
	})
}
