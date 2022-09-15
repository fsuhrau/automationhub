package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var IntroduceProjects *gormigrate.Migration

func init() {
	IntroduceProjects = &gormigrate.Migration{
		ID: "IntroduceProjects",
		Migrate: func(tx *gorm.DB) error {
			// create a new defualt company and move all current data in there
			type Company struct {
				gorm.Model
				Token string
				Name  string
			}
			defaultCompany := Company{
				Token: "default_company",
				Name:  "Default Company",
			}
			if err := tx.Create(&defaultCompany).Error; err != nil {
				return err
			}

			// create project table
			type Project struct {
				gorm.Model
				Identifier string
				CompanyID  uint
				Name       string
				// AccessTokens []*AccessToken
				// Users        []*User `gorm:"many2many:user_projects;"`
				// Apps         []*App
			}
			if err := tx.AutoMigrate(&Project{}); err != nil {
				return err
			}

			// create new default project
			defaultProject := Project{
				Identifier: "default_project",
				CompanyID:  defaultCompany.ID,
				Name:       "Default Project",
			}
			if err := tx.Create(&defaultProject).Error; err != nil {
				return err
			}

			// move accesstokes to project
			if err := tx.Migrator().AddColumn(&models.AccessToken{}, "ProjectID"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&models.AccessToken{}, "CompanyID"); err != nil {
				return err
			}
			if err := tx.Exec("update access_tokens set project_id = ? where id > 0;", defaultProject.ID).Error; err != nil {
				return err
			}

			// rename apps to app binaries
			if err := tx.Migrator().RenameTable("apps", "app_binaries"); err != nil {
				return err
			}
			if err := tx.Migrator().RenameColumn(&models.TestRun{}, "app_id", "app_binary_id"); err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn(&models.AppBinary{}, "AppID"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&models.AppBinary{}, "AppID"); err != nil {
				return err
			}

			// create new app
			type App struct {
				gorm.Model
				ProjectID        uint
				Name             string
				Identifier       string
				Platform         uint
				DefaultParameter string
			}
			if err := tx.AutoMigrate(&App{}); err != nil {
				return err
			}

			// create new default app
			defaultApp := App{
				Identifier: "default_app",
				Platform:   uint(models.PlatformTypeAndroid),
				Name:       "Default App",
				ProjectID:  defaultProject.ID,
			}
			if err := tx.Create(&defaultApp).Error; err != nil {
				return err
			}
			if err := tx.Exec("update app_binaries set app_id = ? where id > 0;", defaultApp.ID).Error; err != nil {
				return err
			}

			if err := tx.Migrator().DropTable("app_functions"); err != nil {
				return err
			}

			if err := tx.Migrator().AddColumn(&models.Test{}, "AppID"); err != nil {
				return err
			}

			type Test struct {
				CompanyID uint
			}

			if err := tx.Migrator().DropColumn(&Test{}, "CompanyID"); err != nil {
				return err
			}

			if err := tx.Exec("update tests set app_id = ? where id > 0;", defaultApp.ID).Error; err != nil {
				return err
			}

			return nil
		},
	}
}
