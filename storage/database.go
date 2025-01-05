package storage

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/migrations"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB

func GetDB(database config.Database) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	checkForInitialSchemaMigration := false
	if _, err := os.Stat(database.SQLiteDBPath); err == nil {
		// old db need to check for initial schema migration
		checkForInitialSchemaMigration = true
	} else if errors.Is(err, os.ErrNotExist) {
		// all fine new db
	} else {
		return nil, err
	}

	var err error
	db, err = gorm.Open(sqlite.Open(database.SQLiteDBPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, err
	}
	tx := db.Begin()

	if checkForInitialSchemaMigration {
		// in case we have already a DB check if we need to apply the initial migration since migrations were introduced later...
		type Migration struct {
			ID string `gorm:"primarykey"`
		}
		if !tx.Migrator().HasTable(&Migration{}) {
			if err = tx.AutoMigrate(&Migration{}); err != nil {
				return nil, err
			}
			if err := tx.Create(&Migration{
				ID: "SCHEMA_INIT",
			}).Error; err != nil {
				return nil, err
			}
		}
	}
	m := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.IntroduceProjects,
		migrations.AddOsInfosToDevice,
		migrations.AddPerformanceAverage,
		migrations.AddNodes,
		migrations.AddUnityTestCategory,
		migrations.AddAliasToDevice,
		migrations.AddPlatformTypeToDevice,
		migrations.AddStatusToNode,
		migrations.AddParentProtocolIdToTestProtocol,
		migrations.AddRoleToUser,
		migrations.AddIndexForUser,
		migrations.RemoveRunAllTestsFromUnityConfig,
	})
	m.InitSchema(migrations.InitSchema)

	if err = m.Migrate(); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return db, nil
}
