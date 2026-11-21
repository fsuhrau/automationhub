package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addAppParameter moves the default parameter of an app into its own table.
func addAppParameter() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddAppParameter", Migrate: migrateAddAppParameter}
}

func migrateAddAppParameter(g *gorm.DB) error {
	type App struct {
		models.Model
	}
	type AppParameter struct {
		models.Model
		AppID uint   `json:"appId"`
		App   *App   `json:"app" gorm:"foreignKey:AppID"`
		Name  string `json:"name"`
		Type  string `json:"type"`
	}
	if err := g.AutoMigrate(&AppParameter{}); err != nil {
		return err
	}
	return g.Migrator().DropColumn(&App{}, "DefaultParameter")
}
