package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddIndexForUser *gormigrate.Migration

func init() {
	AddIndexForUser = &gormigrate.Migration{
		ID: "AddIndexForUser",
		Migrate: func(tx *gorm.DB) error {
			type User struct {
				gorm.Model
				Name string `gorm:"uniqueIndex;not null"`
			}

			if err := tx.Migrator().DropColumn(&User{}, "Name"); err != nil {
				return err
			}

			if err := tx.Migrator().AddColumn(&User{}, "Name"); err != nil {
				return err
			}

			type UserAuth struct {
				gorm.Model
				UserID   uint   `gorm:"uniqueIndex:idx_userid_provider;not null"`
				Email    string `gorm:"uniqueIndex:idx_email_provider;not null"`
				Provider string `gorm:"uniqueIndex:idx_email_provider;uniqueIndex:idx_userid_provider;not null"`
			}

			if err := tx.Migrator().DropColumn(&UserAuth{}, "UserID"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&UserAuth{}, "Email"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&UserAuth{}, "Provider"); err != nil {
				return err
			}

			if err := tx.Migrator().AddColumn(&UserAuth{}, "UserID"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&UserAuth{}, "Email"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&UserAuth{}, "Provider"); err != nil {
				return err
			}
			return nil
		},
	}
}
