package testkit

import (
	"testing"

	"github.com/fsuhrau/automationhub/storage/migrations"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB returns a private in-memory sqlite database with every model migrated,
// the same schema the hub creates. It is closed when the test ends.
func DB(t testing.TB) *gorm.DB {
	t.Helper()
	db, err := openDB("file:"+t.Name()+"?mode=memory&cache=shared", migrations.InitSchema)
	if err != nil {
		t.Fatalf("testkit: open database: %v", err)
	}
	t.Cleanup(func() { closeDB(db) })
	return db
}

func openDB(dsn string, migrate func(*gorm.DB) error) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		closeDB(db)
		return nil, err
	}
	return db, nil
}

func closeDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
