package storage

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/migrations"
	"github.com/glebarez/sqlite"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func getDialect(database config.Database) gorm.Dialector {
	if database.SQLite != nil {
		return sqlite.Open(database.SQLite.Path)
	}
	if database.Postgres != nil {
		return postgres.Open(fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", database.Postgres.User, database.Postgres.Password, database.Postgres.Host+":"+database.Postgres.Port, database.Postgres.Database))
	}

	panic(errors.New("database not found in config"))
	return nil
}

func GetDB(database config.Database) (*gorm.DB, error) {

	if db != nil {
		return db, nil
	}

	var err error
	db, err = gorm.Open(getDialect(database), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, err
	}
	tx := db.Begin()

	m := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{})
	m.InitSchema(migrations.InitSchema)

	if err = m.Migrate(); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return db, nil
}
