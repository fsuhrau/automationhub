package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetDB() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = gorm.Open(sqlite.Open("ah.db"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}