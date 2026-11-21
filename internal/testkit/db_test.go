package testkit

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func TestDBMigratesAllModels(t *testing.T) {
	db := DB(t)
	for _, m := range []any{&models.Project{}, &models.App{}, &models.Test{}, &models.TestRun{}, &models.TestProtocol{},
		&models.ProtocolKey{}, &models.TestRunRecording{}, &models.Device{}, &models.User{}, &models.AccessToken{}, &models.Node{}} {
		if !db.Migrator().HasTable(m) {
			t.Errorf("table for %T missing", m)
		}
	}
}

func TestDBIsIsolatedPerTest(t *testing.T) {
	if err := DB(t).Create(&models.Project{Name: "x"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Run("sub", func(t *testing.T) {
		var n int64
		DB(t).Model(&models.Project{}).Count(&n)
		if n != 0 {
			t.Fatalf("sub test sees %d projects", n)
		}
	})
}

func TestDBClosesOnCleanup(t *testing.T) {
	var db *gorm.DB
	t.Run("inner", func(t *testing.T) { db = DB(t) })
	if err := db.Create(&models.Project{}).Error; err == nil {
		t.Fatal("database still open after cleanup")
	}
}

func TestDBFatalOnOpenError(t *testing.T) {
	expectFatal(t, "x?mode=bogus&", func(tb *fatalTB) { DB(tb) })
}

func TestOpenDBErrors(t *testing.T) {
	if _, err := openDB("file:x?mode=bogus", nil); err == nil {
		t.Error("expected open error")
	}
	_, err := openDB("file:migrate?mode=memory", func(*gorm.DB) error { return errBoom })
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want boom", err)
	}
}
