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
	"time"
)

type Model struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

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

	m := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "AddAppParameter",
			Migrate: func(g *gorm.DB) error {

				type App struct {
					Model
				}

				type AppParameter struct {
					Model
					AppID uint   `json:"appId"`
					App   *App   `json:"app" gorm:"foreignKey:AppID"`
					Name  string `json:"name"`
					Type  string `json:"type"`
				}

				if err := g.AutoMigrate(&AppParameter{}); err != nil {
					return err
				}
				if err := g.Migrator().DropColumn(&App{}, "DefaultParameter"); err != nil {
					return err
				}
				return nil
			},
		},
	})
	m.InitSchema(migrations.InitSchema)

	if err = m.Migrate(); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return db, nil
}
