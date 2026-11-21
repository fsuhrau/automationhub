package storage

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/migrations"
	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"os"
	"time"
)

func getDialect(database config.Database) (gorm.Dialector, error) {
	if database.SQLite != nil {
		return sqlite.Open(database.SQLite.Path), nil
	}
	if database.Postgres != nil {
		return postgres.Open(fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", database.Postgres.User, database.Postgres.Password, database.Postgres.Host+":"+database.Postgres.Port, database.Postgres.Database)), nil
	}

	return nil, errors.New("database not found in config")
}

// newLogger is gorm's default logger without two kinds of noise: a lookup
// that finds nothing (an unknown token, a missing device) is a normal result
// for the caller to handle, not an error, and SQL is logged without its
// values so tokens and passwords never reach the log.
func newLogger(w io.Writer) logger.Interface {
	return logger.New(log.New(w, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  true,
	})
}

// Open connects to the configured database and migrates it to the current
// schema.
func Open(cfg config.Database) (*gorm.DB, error) {
	dialect, err := getDialect(cfg)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dialect, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger(os.Stdout),
	})
	if err != nil {
		return nil, err
	}
	if err := migrations.Run(db); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return db, nil
}

// GetDB is Open under its old name. Deleted with the other aliases by G34.
func GetDB(database config.Database) (*gorm.DB, error) {
	return Open(database)
}
