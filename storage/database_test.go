package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// An existing database gets the personal token columns.
func TestMigrateAddsPersonalTokenColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hub.db")
	old, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type AccessToken struct {
		models.Model
		ProjectID *uint
		NodeID    *uint
		Name      string
		Token     string `gorm:"unique"`
		ExpiresAt *time.Time
	}
	if err := old.AutoMigrate(&AccessToken{}); err != nil {
		t.Fatal(err)
	}
	old.Create(&AccessToken{Name: "ci", Token: "uuid-1"})
	if err := old.Exec("CREATE TABLE migrations (id VARCHAR(255) PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	old.Exec("INSERT INTO migrations (id) VALUES ('SCHEMA_INIT'), ('AddAppParameter')")
	sqlDB, _ := old.DB()
	_ = sqlDB.Close()

	db, err := GetDB(config.Database{SQLite: &config.SQLite{Path: path}})
	if err != nil {
		t.Fatal(err)
	}
	// the upgrade migration ran, not the schema init of a new database
	var ran int64
	db.Table("migrations").Where("id = ?", "AddPersonalAccessTokens").Count(&ran)
	if ran != 1 || db.Migrator().HasTable(&models.Project{}) {
		t.Fatalf("expected only the AddPersonalAccessTokens migration (ran=%d)", ran)
	}
	for _, column := range []string{"user_id", "hint", "last_used_at"} {
		if !db.Migrator().HasColumn(&models.AccessToken{}, column) {
			t.Errorf("column %s missing", column)
		}
	}
	var tok models.AccessToken
	if err := db.First(&tok, "token = ?", "uuid-1").Error; err != nil || tok.UserID != nil {
		t.Fatalf("existing token: %+v %v", tok, err)
	}
}
