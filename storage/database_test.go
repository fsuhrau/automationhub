package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
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

// An existing database gets the info column of bundles, and new and old
// bundles read and write it.
func TestMigrateAddsBundleInfo(t *testing.T) {
	db = nil
	t.Cleanup(func() { db = nil })
	path := filepath.Join(t.TempDir(), "hub.db")
	old, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type AppBinary struct {
		models.Model
		AppID uint
		Name  string
	}
	if err := old.AutoMigrate(&AppBinary{}); err != nil {
		t.Fatal(err)
	}
	old.Create(&AppBinary{AppID: 1, Name: "old"})
	if err := old.Exec("CREATE TABLE migrations (id VARCHAR(255) PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	old.Exec("INSERT INTO migrations (id) VALUES ('SCHEMA_INIT'), ('AddAppParameter'), ('AddPersonalAccessTokens')")
	sqlDB, _ := old.DB()
	_ = sqlDB.Close()

	conn, err := GetDB(config.Database{SQLite: &config.SQLite{Path: path}})
	if err != nil {
		t.Fatal(err)
	}
	if !conn.Migrator().HasColumn(&models.AppBinary{}, "info") {
		t.Fatal("info column missing")
	}
	var existing models.AppBinary
	if err := conn.Select("id", "name", "info").First(&existing).Error; err != nil || existing.Info != nil {
		t.Fatalf("old bundle: %+v %v", existing, err)
	}
	existing.Info = &app.BundleInfo{Format: "apk", Identifier: "com.example"}
	if err := conn.Model(&existing).Select("info").Updates(&existing).Error; err != nil {
		t.Fatal(err)
	}
	var read models.AppBinary
	conn.Select("id", "info").First(&read, existing.ID)
	if read.Info == nil || read.Info.Identifier != "com.example" {
		t.Fatalf("info = %+v", read.Info)
	}
}
