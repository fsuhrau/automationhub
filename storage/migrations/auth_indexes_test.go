package migrations

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"gorm.io/gorm"
)

// legacyUserAuth is user_auths before migration 7: the email index covers empty emails.
type legacyUserAuth struct {
	ID             uint
	UserID         uint   `gorm:"uniqueIndex:idx_userid_provider;not null"`
	Email          string `gorm:"uniqueIndex:idx_email_provider;not null"`
	Provider       string `gorm:"uniqueIndex:idx_email_provider;uniqueIndex:idx_userid_provider;not null"`
	ProviderUserID string
}

func (legacyUserAuth) TableName() string { return "user_auths" }

func authIndexes(t *testing.T, db *gorm.DB) []string {
	t.Helper()
	var names []string
	err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'user_auths' AND name LIKE 'idx_%' AND name NOT LIKE '%deleted_at' ORDER BY name").Scan(&names).Error
	if err != nil {
		t.Fatal(err)
	}
	return names
}

func legacyDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openDB(t)
	before(t, db, &legacyUserAuth{})
	return db
}

func TestPartialAuthIndexesReplaceTheEmailIndex(t *testing.T) {
	db := legacyDB(t)
	rows := []legacyUserAuth{
		{UserID: 1, Email: "a@x", Provider: "password"},
		{UserID: 2, Email: "", Provider: "github", ProviderUserID: "7"},
	}
	db.Create(&rows)
	if got := authIndexes(t, db); !slices.Contains(got, "idx_email_provider") {
		t.Fatalf("old state has no idx_email_provider: %v", got)
	}

	apply(t, migratePartialAuthIndexes, db)
	apply(t, migratePartialAuthIndexes, db) // running again changes nothing

	want := []string{"idx_user_auths_email", "idx_user_auths_identity", "idx_userid_provider"}
	if got := authIndexes(t, db); !slices.Equal(got, want) {
		t.Fatalf("indexes %v, want %v", got, want)
	}
	var n int64
	db.Model(&legacyUserAuth{}).Count(&n)
	if n != 2 {
		t.Fatalf("%d rows survived, want 2", n)
	}
	// accounts without an email no longer collide, a shared email still does
	if err := db.Create(&legacyUserAuth{UserID: 3, Provider: "github", ProviderUserID: "8"}).Error; err != nil {
		t.Fatalf("second account without an email: %v", err)
	}
	if err := db.Create(&legacyUserAuth{UserID: 4, Email: "a@x", Provider: "password"}).Error; err == nil {
		t.Fatal("same email on the same provider was accepted")
	}
	if err := db.Create(&legacyUserAuth{UserID: 5, Provider: "github", ProviderUserID: "7"}).Error; err == nil {
		t.Fatal("same provider user was accepted twice")
	}
}

func TestPartialAuthIndexesRefuseSharedProviderIDs(t *testing.T) {
	db := legacyDB(t)
	db.Create(&[]legacyUserAuth{
		{UserID: 1, Email: "a@x", Provider: "github", ProviderUserID: "7"},
		{UserID: 2, Email: "b@x", Provider: "github", ProviderUserID: "7"},
	})
	err := migratePartialAuthIndexes(db)
	if err == nil || !strings.Contains(err.Error(), `provider "github" share the provider user id "7"`) {
		t.Fatalf("error %v", err)
	}
	var n int64
	db.Model(&legacyUserAuth{}).Count(&n)
	if got := authIndexes(t, db); n != 2 || !slices.Contains(got, "idx_email_provider") {
		t.Fatalf("data changed: %d rows, indexes %v", n, got)
	}
}

func TestPartialAuthIndexesErrors(t *testing.T) {
	tests := map[string]struct {
		prepare func(db *gorm.DB)
		want    string
	}{
		"lookup failure": {func(db *gorm.DB) {
			_ = db.Callback().Row().Before("gorm:row").Register("fail:lookup", func(tx *gorm.DB) {
				if _, grouped := tx.Statement.Clauses["GROUP BY"]; grouped {
					_ = tx.AddError(errors.New("forced failure"))
				}
			})
		}, "look for shared provider user ids"},
		"index failure": {func(db *gorm.DB) { failOn(db, "CREATE UNIQUE INDEX") }, "CREATE UNIQUE INDEX"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db := openDB(t)
			before(t, db, &legacyUserAuth{})
			tt.prepare(db)
			if err := migratePartialAuthIndexes(db); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %v, want %q", err, tt.want)
			}
		})
	}
}

// A new database gets the same indexes from InitSchema as an old one from migration 7.
func TestFreshSchemaHasTheAuthIndexes(t *testing.T) {
	db := openDB(t)
	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	want := []string{"idx_user_auths_email", "idx_user_auths_identity", "idx_userid_provider"}
	if got := authIndexes(t, db); !slices.Equal(got, want) {
		t.Fatalf("indexes %v, want %v", got, want)
	}
	var sql []string
	db.Raw("SELECT sql FROM sqlite_master WHERE name IN ('idx_user_auths_email','idx_user_auths_identity') ORDER BY name").Scan(&sql)
	if len(sql) != 2 || !strings.Contains(sql[0], "(`email`,`provider`) WHERE email <> ''") || !strings.Contains(sql[1], "(`provider`,`provider_user_id`) WHERE provider_user_id <> ''") {
		t.Fatalf("index sql %q", sql)
	}
}

// A database without the accounts table has nothing to index.
func TestPartialAuthIndexesWithoutTheTable(t *testing.T) {
	apply(t, migratePartialAuthIndexes, openDB(t))
}
