package persistence

import (
	"github.com/fsuhrau/automationhub/persistence/models"
	"gorm.io/gorm"
)

func migrate(db *gorm.DB) error {
	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.App{})
	db.AutoMigrate(&models.Test{})
	db.AutoMigrate(&models.Device{})
	return nil
}