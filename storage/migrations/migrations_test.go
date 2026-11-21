package migrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openDB is a private in-memory database. It does not use internal/testkit,
// because testkit imports this package.
func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return db
}

// before creates the table in the state before a migration.
func before(t *testing.T, db *gorm.DB, table any) {
	t.Helper()
	if err := db.AutoMigrate(table); err != nil {
		t.Fatal(err)
	}
}

// failOn makes every statement containing sql fail.
func failOn(db *gorm.DB, sql string) {
	_ = db.Callback().Raw().Before("gorm:raw").Register("fail:"+sql, func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), sql) {
			_ = tx.AddError(errors.New("forced failure"))
		}
	})
}

func apply(t *testing.T, m func(*gorm.DB) error, db *gorm.DB) {
	t.Helper()
	if err := m(db); err != nil {
		t.Fatal(err)
	}
}

func TestAllIDsAreUniqueAndOrdered(t *testing.T) {
	want := []string{"AddAppParameter", "AddPersonalAccessTokens", "AddBundleInfo", "AddBundleArchivedAt", "AddSessionRecording", "AddProtocolKeys"}
	all := All()
	if len(all) != len(want) {
		t.Fatalf("got %d migrations, want %d", len(all), len(want))
	}
	for i, m := range all {
		if m.ID != want[i] || m.Migrate == nil {
			t.Errorf("migration %d = %q, want %q", i, m.ID, want[i])
		}
	}
}

func TestAddAppParameter(t *testing.T) {
	db := openDB(t)
	type App struct {
		models.Model
		DefaultParameter string
	}
	before(t, db, &App{})
	apply(t, migrateAddAppParameter, db)
	// DropColumn gets the name "DefaultParameter", which matches no column
	// (they are snake case), so default_parameter stays. Behaviour kept from
	// before the split; see the report.
	if !db.Migrator().HasTable("app_parameters") {
		t.Fatal("app_parameters table missing")
	}
	for _, c := range []string{"app_id", "name", "type"} {
		if !db.Migrator().HasColumn("app_parameters", c) {
			t.Errorf("column %s missing", c)
		}
	}
}

func TestAddPersonalAccessTokens(t *testing.T) {
	db := openDB(t)
	type AccessToken struct {
		models.Model
		Token string
	}
	before(t, db, &AccessToken{})
	apply(t, migrateAddPersonalAccessTokens, db)
	for _, c := range []string{"user_id", "hint", "last_used_at"} {
		if !db.Migrator().HasColumn("access_tokens", c) {
			t.Errorf("column %s missing", c)
		}
	}
}

func TestAddBundleColumns(t *testing.T) {
	type AppBinary struct {
		models.Model
		Name string
	}
	tests := []struct {
		name    string
		migrate func(*gorm.DB) error
		column  string
	}{
		{"info", migrateAddBundleInfo, "info"},
		{"archived at", migrateAddBundleArchivedAt, "archived_at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openDB(t)
			before(t, db, &AppBinary{})
			apply(t, tt.migrate, db)
			if !db.Migrator().HasColumn("app_binaries", tt.column) {
				t.Fatalf("column %s missing", tt.column)
			}
			apply(t, tt.migrate, db) // rerun is harmless
		})
		t.Run(tt.name+" without table", func(t *testing.T) {
			db := openDB(t)
			apply(t, tt.migrate, db)
			if db.Migrator().HasTable("app_binaries") {
				t.Fatal("table must not be created")
			}
		})
	}
}

func TestAddSessionRecording(t *testing.T) {
	db := openDB(t)
	type TestConfig struct {
		models.Model
		TestID uint
	}
	before(t, db, &TestConfig{})
	apply(t, migrateAddSessionRecording, db)
	if !db.Migrator().HasColumn("test_configs", "record_session") || !db.Migrator().HasTable(&models.TestRunRecording{}) {
		t.Fatal("expected record_session column and recordings table")
	}
}

func TestAddSessionRecordingWithoutConfigTable(t *testing.T) {
	db := openDB(t)
	apply(t, migrateAddSessionRecording, db)
	if db.Migrator().HasTable("test_configs") || !db.Migrator().HasTable(&models.TestRunRecording{}) {
		t.Fatal("expected only the recordings table")
	}
}

func TestAddProtocolKeys(t *testing.T) {
	db := openDB(t)
	apply(t, migrateAddProtocolKeys, db)
	if !db.Migrator().HasTable(&models.ProtocolKey{}) {
		t.Fatal("protocol keys table missing")
	}
}

func TestRunNewDatabaseGetsInitSchemaOnly(t *testing.T) {
	db := openDB(t)
	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	var ran int64
	db.Table("migrations").Count(&ran)
	if !db.Migrator().HasTable(&models.Project{}) || ran != int64(len(All())+1) {
		t.Fatalf("project table=%v, recorded=%d", db.Migrator().HasTable(&models.Project{}), ran)
	}
}

func TestRunFailureRollsBack(t *testing.T) {
	db := openDB(t)
	db.Exec("CREATE TABLE migrations (other VARCHAR(255))")
	if err := Run(db); err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrationErrorsAreReturned(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, db *gorm.DB)
		fail    string
		migrate func(*gorm.DB) error
	}{
		{"app parameter create", func(t *testing.T, db *gorm.DB) {}, "CREATE TABLE", migrateAddAppParameter},
		{"personal tokens", func(t *testing.T, db *gorm.DB) {}, "CREATE TABLE", migrateAddPersonalAccessTokens},
		{"session recording column", func(t *testing.T, db *gorm.DB) {
			type TestConfig struct{ models.Model }
			before(t, db, &TestConfig{})
		}, "ALTER TABLE", migrateAddSessionRecording},
		{"bundle info column", func(t *testing.T, db *gorm.DB) {
			type AppBinary struct{ models.Model }
			before(t, db, &AppBinary{})
		}, "ALTER TABLE", migrateAddBundleInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openDB(t)
			tt.prepare(t, db)
			failOn(db, tt.fail)
			if err := tt.migrate(db); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestInitSchemaCreatesEveryModelTable(t *testing.T) {
	db := openDB(t)
	apply(t, InitSchema, db)
	for _, m := range schemaModels() {
		if !db.Migrator().HasTable(m) {
			t.Errorf("table of %T missing", m)
		}
	}
}

func TestInitSchemaReturnsError(t *testing.T) {
	db := openDB(t)
	failOn(db, "CREATE TABLE")
	if err := InitSchema(db); err == nil {
		t.Fatal("expected error")
	}
}
