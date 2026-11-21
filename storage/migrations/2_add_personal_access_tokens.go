package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addPersonalAccessTokens gives access tokens an owner, a hint and a last use.
func addPersonalAccessTokens() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddPersonalAccessTokens", Migrate: migrateAddPersonalAccessTokens}
}

func migrateAddPersonalAccessTokens(g *gorm.DB) error {
	type AccessToken struct {
		UserID     *uint `gorm:"default:NULL;index"`
		Hint       string
		LastUsedAt *time.Time
	}
	return g.AutoMigrate(&AccessToken{})
}
