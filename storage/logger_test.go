package storage

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// A lookup that finds nothing isn't logged, and logged SQL has no values.
func TestLoggerSkipsNotFoundAndHidesValues(t *testing.T) {
	var out bytes.Buffer
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{Logger: newLogger(&out)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.AccessToken{}); err != nil {
		t.Fatal(err)
	}

	var token models.AccessToken
	if err := db.First(&token, "token = ?", "secret-1235123").Error; err != gorm.ErrRecordNotFound {
		t.Fatalf("expected record not found, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("not found was logged: %q", out.String())
	}

	// a real error is still logged, without the value
	db.Exec("SELECT * FROM missing_table WHERE token = ?", "secret-1235123")
	if !strings.Contains(out.String(), "missing_table") {
		t.Fatalf("error was not logged: %q", out.String())
	}
	if strings.Contains(out.String(), "secret-1235123") {
		t.Fatalf("value leaked into the log: %q", out.String())
	}
}
