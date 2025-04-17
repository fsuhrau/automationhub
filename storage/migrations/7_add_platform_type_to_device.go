package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddPlatformTypeToDevice *gormigrate.Migration

func init() {
	AddPlatformTypeToDevice = &gormigrate.Migration{
		ID: "AddPlatformTypeToDevice",
		Migrate: func(tx *gorm.DB) error {
			type Device struct {
				gorm.Model
				PlatformType uint
			}

			if err := tx.Migrator().AddColumn(&Device{}, "PlatformType"); err != nil {
				return err
			}

			if err := tx.Exec("update devices set platform_type = 0 where manager = 'ios_device';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 1 where manager = 'android_device';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 2 where manager = 'mac';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 3 where manager = 'windows';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 4 where manager = 'linux';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 5 where manager = 'web';").Error; err != nil {
				return err
			}
			if err := tx.Exec("update devices set platform_type = 6 where manager = 'unity_editor';").Error; err != nil {
				return err
			}

			return nil
		},
	}
}
