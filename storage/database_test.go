package storage

import (
	"errors"
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

	db, err := Open(config.Database{SQLite: &config.SQLite{Path: path}})
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

	conn, err := Open(config.Database{SQLite: &config.SQLite{Path: path}})
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

// Existing hubs get the record option on their tests (off) and the table of
// recordings.
func TestMigrateAddsSessionRecording(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hub.db")
	old, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type TestConfig struct {
		models.Model
		TestID     uint
		AllDevices bool
	}
	if err := old.AutoMigrate(&TestConfig{}); err != nil {
		t.Fatal(err)
	}
	old.Create(&TestConfig{TestID: 1, AllDevices: true})
	if err := old.Exec("CREATE TABLE migrations (id VARCHAR(255) PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	old.Exec("INSERT INTO migrations (id) VALUES ('SCHEMA_INIT'), ('AddAppParameter'), ('AddPersonalAccessTokens'), ('AddBundleInfo'), ('AddBundleArchivedAt')")
	sqlDB, _ := old.DB()
	_ = sqlDB.Close()

	conn, err := Open(config.Database{SQLite: &config.SQLite{Path: path}})
	if err != nil {
		t.Fatal(err)
	}
	var existing models.TestConfig
	if err := conn.Select("id", "record_session").First(&existing).Error; err != nil || existing.RecordSession {
		t.Fatalf("old config: %+v %v", existing, err)
	}
	if !conn.Migrator().HasTable(&models.TestRunRecording{}) {
		t.Fatal("recordings table missing")
	}
}

func sqliteConfig(t *testing.T) config.Database {
	return config.Database{SQLite: &config.SQLite{Path: filepath.Join(t.TempDir(), "hub.db")}}
}

// A new database gets every model's table with its key columns.
func TestOpenCreatesFullSchema(t *testing.T) {
	db, err := Open(sqliteConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	wantColumns := []struct {
		model   any
		columns []string
	}{
		{&models.Company{}, []string{"id", "name"}},
		{&models.Project{}, []string{"id", "name", "company_id"}},
		{&models.AccessToken{}, []string{"id", "token", "user_id", "hint", "last_used_at", "project_id", "expires_at"}},
		{&models.User{}, []string{"id", "name", "role"}},
		{&models.UserAuth{}, []string{"id"}},
		{&models.App{}, []string{"id", "name", "project_id"}},
		{&models.AppParameter{}, []string{"id", "app_id", "name", "type"}},
		{&models.AppBinary{}, []string{"id", "app_id", "name", "info", "archived_at"}},
		{&models.Node{}, []string{"id"}},
		{&models.Device{}, []string{"id", "device_identifier", "name", "manager", "os", "os_version"}},
		{&models.ConnectionParameter{}, []string{"id", "device_id", "ip", "port"}},
		{&models.DeviceParameter{}, []string{"id", "device_id", "key", "value"}},
		{&models.CustomParameter{}, []string{"id", "device_id", "key", "value"}},
		{&models.DeviceLog{}, []string{"id", "device_id", "status", "payload"}},
		{&models.Test{}, []string{"id", "app_id", "name"}},
		{&models.TestConfig{}, []string{"id", "test_id", "record_session"}},
		{&models.TestConfigDevice{}, []string{"id"}},
		{&models.TestConfigScenario{}, []string{"id"}},
		{&models.ScenarioStep{}, []string{"id"}},
		{&models.TestConfigUnity{}, []string{"id"}},
		{&models.UnityTestFunction{}, []string{"id"}},
		{&models.TestProtocol{}, []string{"id"}},
		{&models.ProtocolEntry{}, []string{"id"}},
		{&models.ProtocolPerformanceEntry{}, []string{"id"}},
		{&models.TestRun{}, []string{"id", "test_id"}},
		{&models.TestRunLogEntry{}, []string{"id"}},
		{&models.TestRunDeviceStatus{}, []string{"id"}},
		{&models.TestRunRecording{}, []string{"id"}},
		{&models.ProtocolKey{}, []string{"id"}},
	}
	for _, w := range wantColumns {
		if !db.Migrator().HasTable(w.model) {
			t.Errorf("table of %T missing", w.model)
			continue
		}
		for _, c := range w.columns {
			if !db.Migrator().HasColumn(w.model, c) {
				t.Errorf("%T: column %s missing", w.model, c)
			}
		}
	}
}

func TestOpenErrors(t *testing.T) {
	broken := sqliteConfig(t)
	seed, err := gorm.Open(sqlite.Open(broken.SQLite.Path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// a migrations table without id makes the migration runner fail
	seed.Exec("CREATE TABLE migrations (other VARCHAR(255))")
	sqlDB, _ := seed.DB()
	_ = sqlDB.Close()

	tests := []struct {
		name string
		cfg  config.Database
	}{
		{"no database configured", config.Database{}},
		{"postgres unreachable", config.Database{Postgres: &config.Postgres{Host: "127.0.0.1", Port: "1", User: "u", Database: "d"}}},
		{"migration fails", broken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if db, err := Open(tt.cfg); err == nil || db != nil {
				t.Fatalf("db=%v err=%v, want error", db, err)
			}
		})
	}
}

// Open translates a duplicate key into gorm.ErrDuplicatedKey, which
// storeerr.Map turns into a conflict (409) instead of an internal error.
func TestOpenTranslatesADuplicateKey(t *testing.T) {
	db, err := Open(config.Database{SQLite: &config.SQLite{Path: filepath.Join(t.TempDir(), "hub.db")}})
	if err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Node{Identifier: "n1", Name: "one"})

	err = db.Create(&models.Node{Identifier: "n1", Name: "two"}).Error

	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		t.Fatalf("err = %v, want gorm.ErrDuplicatedKey", err)
	}
}
