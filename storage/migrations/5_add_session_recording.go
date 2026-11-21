package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addSessionRecording adds the record option to test configs and the table of
// recordings.
func addSessionRecording() *gormigrate.Migration {
	return &gormigrate.Migration{ID: "AddSessionRecording", Migrate: migrateAddSessionRecording}
}

func migrateAddSessionRecording(g *gorm.DB) error {
	type TestConfig struct {
		RecordSession bool
	}
	if err := addColumnIfMissing(g, &TestConfig{}, "RecordSession"); err != nil {
		return err
	}
	return g.AutoMigrate(&models.TestRunRecording{})
}
