package projectstore_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/projectstore"
	"gorm.io/gorm"
)

func setup(t *testing.T) (*projectstore.Store, *gorm.DB) {
	t.Helper()
	db := testkit.DB(t)
	return projectstore.New(db), db
}

// closed is a store whose database is gone, to reach the error paths.
func closed(t *testing.T) *projectstore.Store {
	t.Helper()
	store, db := setup(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()
	return store
}

func TestListLoadsAppsAndParameters(t *testing.T) {
	store, db := setup(t)
	app := testkit.App(db)
	db.Create(&models.AppParameter{AppID: app.ID, Name: "level", Type: "string"})
	got, err := store.List(context.Background())
	if err != nil || len(got) != 1 || len(got[0].Apps) != 1 || len(got[0].Apps[0].Parameter) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCreateUniqueSkipsTakenIdentifiersAlsoDeleted(t *testing.T) {
	store, db := setup(t)
	gone := models.Project{Name: "Game", Identifier: "game-2"}
	db.Create(&gone)
	db.Delete(&gone)
	db.Create(&models.Project{Name: "Game", Identifier: "game"})

	p := &models.Project{Name: "Game"}
	if err := store.CreateUnique(context.Background(), p, "game"); err != nil || p.Identifier != "game-3" || p.ID == 0 {
		t.Fatalf("got %+v, %v", p, err)
	}
	fresh := &models.Project{Name: "Other"}
	if err := store.CreateUnique(context.Background(), fresh, "other"); err != nil || fresh.Identifier != "other" {
		t.Fatalf("got %+v, %v", fresh, err)
	}
}

func TestByIdentifierAndSave(t *testing.T) {
	store, db := setup(t)
	testkit.Project(db, func(p *models.Project) { p.Identifier = "proj" })
	p, err := store.ByIdentifier(context.Background(), "proj")
	if err != nil {
		t.Fatal(err)
	}
	p.Name = "Renamed"
	if err := store.Save(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	var stored models.Project
	db.First(&stored, p.ID)
	if stored.Name != "Renamed" {
		t.Fatalf("stored %+v", stored)
	}
	if _, err := store.ByIdentifier(context.Background(), "nope"); apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("unknown: %v", err)
	}
}

func TestDelete(t *testing.T) {
	store, db := setup(t)
	p := testkit.Project(db)
	if err := store.Delete(context.Background(), p.ID); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&models.Project{}).Where("id = ?", p.ID).Count(&n)
	if n != 0 {
		t.Fatal("project still there")
	}
	if err := store.Delete(context.Background(), p.ID); apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("again: %v", err)
	}
}

func TestDatabaseErrorsAreInternal(t *testing.T) {
	store := closed(t)
	ctx := context.Background()
	p := &models.Project{Name: "x", Identifier: "x"}
	calls := map[string]error{
		"list":   func() error { _, err := store.List(ctx); return err }(),
		"create": store.CreateUnique(ctx, p, "x"),
		"by id":  func() error { _, err := store.ByIdentifier(ctx, "x"); return err }(),
		"save":   store.Save(ctx, p),
		"delete": store.Delete(ctx, 1),
	}
	for name, err := range calls {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestCreateUniqueFailsWhenTheLookupFails(t *testing.T) {
	store, db := setup(t)
	if err := db.Migrator().DropTable(&models.Project{}); err != nil {
		t.Fatal(err)
	}
	err := store.CreateUnique(context.Background(), &models.Project{Name: "x"}, "x")
	if apperr.KindOf(err) != apperr.KindInternal {
		t.Fatalf("err %v", err)
	}
}
