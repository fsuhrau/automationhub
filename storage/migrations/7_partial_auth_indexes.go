package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// partialAuthIndexes lets accounts without an email coexist and makes the
// provider's user id unique.
func partialAuthIndexes() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "PartialAuthIndexes", Migrate: migratePartialAuthIndexes}
}

// authIndexStatements are valid SQL for SQLite and PostgreSQL and can run again.
func authIndexStatements() []string {
	return []string{
		"DROP INDEX IF EXISTS idx_email_provider",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_auths_email ON user_auths (email, provider) WHERE email <> ''",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_auths_identity ON user_auths (provider, provider_user_id) WHERE provider_user_id <> ''",
	}
}

func migratePartialAuthIndexes(g *gorm.DB) error {
	if !g.Migrator().HasTable("user_auths") {
		return nil // nothing to index; InitSchema or a later AutoMigrate creates the table with the indexes
	}
	if err := refuseSharedProviderIDs(g); err != nil {
		return err
	}
	for _, statement := range authIndexStatements() {
		if err := g.Exec(statement).Error; err != nil {
			return fmt.Errorf("%s: %w", statement, err)
		}
	}
	return nil
}

// refuseSharedProviderIDs fails when two accounts claim the same provider user:
// the migration never deletes accounts, someone has to resolve them first.
func refuseSharedProviderIDs(g *gorm.DB) error {
	var shared struct{ Provider, ProviderUserID string }
	err := g.Table("user_auths").Select("provider, provider_user_id").
		Where("provider_user_id <> ''").Group("provider, provider_user_id").
		Having("COUNT(*) > 1").Limit(1).Scan(&shared).Error
	if err != nil {
		return fmt.Errorf("look for shared provider user ids: %w", err)
	}
	if shared.Provider != "" || shared.ProviderUserID != "" {
		return fmt.Errorf("several accounts of provider %q share the provider user id %q: merge or delete the duplicates in user_auths, then start again", shared.Provider, shared.ProviderUserID)
	}
	return nil
}
