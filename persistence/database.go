package persistence

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Get() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = gorm.Open(sqlite.Open("ah.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}