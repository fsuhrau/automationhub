package appstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/appstore"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// fixture is a store on two projects, each with one app.
type fixture struct {
	db        *gorm.DB
	store     *appstore.Store
	mine, foe *models.App
	sc, other appsvc.Scope
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := testkit.DB(t)
	mine, foe := testkit.App(db), testkit.App(db)
	return &fixture{
		db: db, store: appstore.New(db), mine: mine, foe: foe,
		sc:    appsvc.Scope{ProjectID: mine.ProjectID, AppID: mine.ID},
		other: appsvc.Scope{ProjectID: foe.ProjectID, AppID: mine.ID},
	}
}

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != kind {
		t.Fatalf("error %v has kind %d, want %d", err, got, kind)
	}
}

func TestAppsAreScopedToTheProjectAndHaveTheirParameters(t *testing.T) {
	f := newFixture(t)
	f.db.Create(&models.AppParameter{AppID: f.mine.ID, Name: "level", Type: "string"})
	ctx := context.Background()

	got, err := f.store.Apps(ctx, f.mine.ProjectID)
	if err != nil || len(got) != 1 || got[0].ID != f.mine.ID || len(got[0].Parameter) != 1 {
		t.Fatalf("apps %+v, %v", got, err)
	}
	one, err := f.store.App(ctx, f.sc)
	if err != nil || one.ID != f.mine.ID || len(one.Parameter) != 1 {
		t.Fatalf("app %+v, %v", one, err)
	}
	_, err = f.store.App(ctx, f.other)
	wantKind(t, err, apperr.KindNotFound)
}

func TestCreateAndSaveApp(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	created := models.App{ProjectID: f.mine.ProjectID, Name: "New", Identifier: "com.new"}
	if err := f.store.CreateApp(ctx, &created); err != nil || created.ID == 0 {
		t.Fatalf("create: %+v, %v", created, err)
	}
	created.Name = "Renamed"
	created.Parameter = []models.AppParameter{{Name: "ignored", Type: "string"}}
	if err := f.store.SaveApp(ctx, &created); err != nil {
		t.Fatal(err)
	}
	var stored models.App
	f.db.Preload("Parameter").First(&stored, created.ID)
	if stored.Name != "Renamed" || len(stored.Parameter) != 0 {
		t.Errorf("stored %+v, want the new name and no parameters written", stored)
	}
}

func TestParameters(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	p := models.AppParameter{AppID: f.mine.ID, Name: "level", Type: "string"}
	if err := f.store.CreateParameter(ctx, &p); err != nil || p.ID == 0 {
		t.Fatalf("create: %v", err)
	}
	if got, err := f.store.Parameters(ctx, f.sc); err != nil || len(got) != 1 {
		t.Fatalf("parameters %+v, %v", got, err)
	}
	if got, err := f.store.Parameters(ctx, f.other); err != nil || len(got) != 0 {
		t.Fatalf("another project's parameters %+v, %v; want none", got, err)
	}
	ref := appsvc.ParameterRef{Scope: f.sc, ID: p.ID}
	found, err := f.store.Parameter(ctx, ref)
	if err != nil || found.Name != "level" {
		t.Fatalf("parameter %+v, %v", found, err)
	}
	_, err = f.store.Parameter(ctx, appsvc.ParameterRef{Scope: f.other, ID: p.ID})
	wantKind(t, err, apperr.KindNotFound)

	found.Name = "stage"
	if err := f.store.SaveParameter(ctx, found); err != nil {
		t.Fatal(err)
	}
	if err := f.store.DeleteParameter(ctx, appsvc.ParameterRef{Scope: f.other, ID: p.ID}); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.store.Parameters(ctx, f.sc); len(got) != 1 || got[0].Name != "stage" {
		t.Fatalf("after another project's delete and a save: %+v", got)
	}
	if err := f.store.DeleteParameter(ctx, ref); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.store.Parameters(ctx, f.sc); len(got) != 0 {
		t.Errorf("parameters left: %+v", got)
	}
}

func TestBundles(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	older := &models.AppBinary{AppID: f.mine.ID, Version: "1", Hash: "h1"}
	newer := &models.AppBinary{AppID: f.mine.ID, Version: "2", Hash: "h2"}
	for _, b := range []*models.AppBinary{older, newer} {
		if err := f.store.CreateBundle(ctx, b); err != nil || b.ID == 0 {
			t.Fatalf("create: %v", err)
		}
	}
	f.db.Create(&models.AppBinary{AppID: f.foe.ID, Hash: "h1"})

	list, err := f.store.Bundles(ctx, f.sc)
	if err != nil || len(list) != 2 || list[0].ID != newer.ID {
		t.Fatalf("bundles %+v, %v; want the newest first", list, err)
	}
	if list, _ := f.store.Bundles(ctx, f.other); len(list) != 0 {
		t.Errorf("another project's bundles: %+v", list)
	}
	byHash, err := f.store.BundleByHash(ctx, f.sc, "h1")
	if err != nil || byHash.ID != older.ID {
		t.Fatalf("by hash %+v, %v; want this app's bundle", byHash, err)
	}
	_, err = f.store.BundleByHash(ctx, f.sc, "none")
	wantKind(t, err, apperr.KindNotFound)
	_, err = f.store.BundleByHash(ctx, f.other, "h1")
	wantKind(t, err, apperr.KindNotFound)
}

func TestBundleLookupAndWrites(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	b := &models.AppBinary{AppID: f.mine.ID, Name: "old", Version: "1", Hash: "h"}
	f.db.Create(b)
	ref := appsvc.BundleRef{Scope: f.sc, ID: b.ID}

	found, err := f.store.Bundle(ctx, ref)
	if err != nil || found.ID != b.ID {
		t.Fatalf("bundle %+v, %v", found, err)
	}
	for _, bad := range []appsvc.BundleRef{{Scope: f.other, ID: b.ID}, {Scope: f.sc, ID: 999}} {
		_, err := f.store.Bundle(ctx, bad)
		wantKind(t, err, apperr.KindNotFound)
	}

	found.Tags = "beta"
	if err := f.store.SaveTags(ctx, found); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	found.ArchivedAt = &now
	if err := f.store.SaveArchivedAt(ctx, found); err != nil {
		t.Fatal(err)
	}
	found.Name, found.Version = "new", "ignored"
	if err := f.store.SaveRead(ctx, found, []string{"name"}); err != nil {
		t.Fatal(err)
	}
	var stored models.AppBinary
	f.db.First(&stored, b.ID)
	if stored.Tags != "beta" || !stored.Archived() || stored.Name != "new" || stored.Version != "1" {
		t.Errorf("stored %+v, want tags, archive time and the read name only", stored)
	}

	found.ArchivedAt = nil
	if err := f.store.SaveArchivedAt(ctx, found); err != nil {
		t.Fatal(err)
	}
	var restored models.AppBinary
	f.db.First(&restored, b.ID)
	if restored.Archived() {
		t.Error("a nil archive time must bring the bundle back")
	}
	if err := f.store.DeleteBundle(ctx, found); err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.Bundle(ctx, ref); err == nil {
		t.Error("the bundle must be gone")
	}
}

func TestRuns(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	busy := &models.AppBinary{AppID: f.mine.ID, Hash: "a"}
	idle := &models.AppBinary{AppID: f.mine.ID, Hash: "b"}
	f.db.Create(busy)
	f.db.Create(idle)
	test := testkit.Test(f.db, func(x *models.Test) { x.AppID, x.Name = f.mine.ID, "Smoke" })
	for i := 0; i < 3; i++ {
		testkit.TestRun(f.db, func(r *models.TestRun) { r.TestID, r.AppBinaryID = test.ID, busy.ID })
	}

	counts, err := f.store.RunCounts(ctx, []uint{busy.ID, idle.ID})
	if err != nil || counts[busy.ID] != 3 || counts[idle.ID] != 0 {
		t.Fatalf("counts %v, %v", counts, err)
	}
	if counts, err := f.store.RunCounts(ctx, nil); err != nil || len(counts) != 0 {
		t.Fatalf("counts of no bundles: %v, %v", counts, err)
	}
	if n, err := f.store.RunCount(ctx, busy.ID); err != nil || n != 3 {
		t.Fatalf("count %d, %v", n, err)
	}
	recent, err := f.store.RecentRuns(ctx, busy.ID, 2)
	if err != nil || len(recent) != 2 || recent[0].ID < recent[1].ID || recent[0].Test == nil || recent[0].Test.Name != "Smoke" {
		t.Fatalf("recent %+v, %v; want the newest two with their test", recent, err)
	}
	ids, err := f.store.RecentRunIDs(ctx, busy.ID, 100)
	if err != nil || len(ids) != 3 || ids[0] < ids[2] {
		t.Fatalf("ids %v, %v", ids, err)
	}
}

func TestStoreErrorsAreMapped(t *testing.T) {
	f := newFixture(t)
	sqlDB, _ := f.db.DB()
	_ = sqlDB.Close()
	ctx := context.Background()
	calls := map[string]error{}
	_, calls["Apps"] = f.store.Apps(ctx, 1)
	_, calls["App"] = f.store.App(ctx, f.sc)
	calls["CreateApp"] = f.store.CreateApp(ctx, &models.App{})
	calls["SaveApp"] = f.store.SaveApp(ctx, &models.App{})
	_, calls["Parameters"] = f.store.Parameters(ctx, f.sc)
	_, calls["Parameter"] = f.store.Parameter(ctx, appsvc.ParameterRef{Scope: f.sc, ID: 1})
	calls["CreateParameter"] = f.store.CreateParameter(ctx, &models.AppParameter{})
	calls["SaveParameter"] = f.store.SaveParameter(ctx, &models.AppParameter{})
	calls["DeleteParameter"] = f.store.DeleteParameter(ctx, appsvc.ParameterRef{Scope: f.sc, ID: 1})
	_, calls["Bundles"] = f.store.Bundles(ctx, f.sc)
	_, calls["Bundle"] = f.store.Bundle(ctx, appsvc.BundleRef{Scope: f.sc, ID: 1})
	_, calls["BundleByHash"] = f.store.BundleByHash(ctx, f.sc, "h")
	calls["CreateBundle"] = f.store.CreateBundle(ctx, &models.AppBinary{})
	calls["DeleteBundle"] = f.store.DeleteBundle(ctx, &models.AppBinary{})
	calls["SaveTags"] = f.store.SaveTags(ctx, &models.AppBinary{})
	calls["SaveArchivedAt"] = f.store.SaveArchivedAt(ctx, &models.AppBinary{})
	calls["SaveRead"] = f.store.SaveRead(ctx, &models.AppBinary{}, []string{"name"})
	_, calls["RunCounts"] = f.store.RunCounts(ctx, []uint{1})
	_, calls["RunCount"] = f.store.RunCount(ctx, 1)
	_, calls["RecentRuns"] = f.store.RecentRuns(ctx, 1, 1)
	_, calls["RecentRunIDs"] = f.store.RecentRunIDs(ctx, 1, 1)
	for name, err := range calls {
		if err == nil {
			t.Errorf("%s: a closed database must fail", name)
			continue
		}
		wantKind(t, err, apperr.KindInternal)
	}
}
