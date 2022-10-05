package storage

import (
	"github.com/fsuhrau/automationhub/storage/migrations"
	"github.com/go-gormigrate/gormigrate/v2"
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
	})

	if err != nil {
		return nil, err
	}
	tx := db.Begin()
	m := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// migrations.InitialMigration,
		migrations.IntroduceProjects,
		migrations.AddOsInfosToDevice,
		migrations.AddPerformanceAverage,
	})

	if err = m.Migrate(); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return db, nil
}
