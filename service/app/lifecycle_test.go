package app

import (
	"context"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestDeleteBundleRemovesItsFileAndIcon(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{AppPath: "game_abc.apk"})
	other := f.addBundle(models.AppBinary{AppPath: "other.apk"})
	f.files.data["/store/game_abc.apk"] = []byte("x")
	f.files.data["/store/game_abc.apk.icon"] = []byte("i")
	f.files.data["/store/other.apk"] = []byte("x")

	if err := f.svc.DeleteBundle(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	if f.files.has("/store/game_abc.apk") || f.files.has("/store/game_abc.apk.icon") || !f.files.has("/store/other.apk") {
		t.Errorf("files left: %v", f.files.data)
	}
	if len(f.repo.bundles) != 1 || f.repo.bundles[0].ID != other.ID {
		t.Errorf("bundles left: %+v", f.repo.bundles)
	}
}

func TestDeleteBundleWithoutAFilePathTouchesNoFile(t *testing.T) {
	for _, path := range []string{"", "../outside.txt"} {
		f := newFixture(t)
		ref := f.addBundle(models.AppBinary{AppPath: path})
		f.files.data["/store/keep"] = []byte("x")
		if err := f.svc.DeleteBundle(context.Background(), ref); err != nil {
			t.Fatal(err)
		}
		if !f.files.has("/store/keep") || len(f.repo.bundles) != 0 {
			t.Errorf("path %q: files %v, bundles %+v", path, f.files.data, f.repo.bundles)
		}
	}
}

func TestDeleteBundleErrors(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{AppPath: "a.apk"})
	wantKind(t, f.svc.DeleteBundle(context.Background(), BundleRef{Scope: f.sc, ID: 99}), apperr.KindNotFound)
	wantKind(t, f.svc.DeleteBundle(context.Background(), BundleRef{Scope: Scope{ProjectID: 2, AppID: 1}, ID: ref.ID}), apperr.KindNotFound)

	f.files.fail["Remove:/store/a.apk"] = errBoom
	if err := f.svc.DeleteBundle(context.Background(), ref); err == nil || len(f.repo.bundles) != 1 {
		t.Fatalf("a file that can't go must keep the bundle: %v", err)
	}
	delete(f.files.fail, "Remove:/store/a.apk")
	f.repo.fail["DeleteBundle"] = errBoom
	if err := f.svc.DeleteBundle(context.Background(), ref); err == nil {
		t.Fatal("a store error must come back")
	}
}

func archivable(f *fixture) BundleRef {
	f.analyser.icon, f.analyser.iconType = []byte("png"), "image/png"
	f.files.data["/store/a.apk"] = []byte("x")
	return f.addBundle(models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{Icon: "icon.png"}})
}

func TestArchiveKeepsTheIconAndRemovesTheFile(t *testing.T) {
	f := newFixture(t)
	ref := archivable(f)
	got, err := f.svc.Archive(context.Background(), ref)
	if err != nil || !got.Archived() {
		t.Fatalf("archive %+v, %v", got, err)
	}
	if f.files.has("/store/a.apk") || string(f.files.data["/store/a.apk.icon"]) != "png" {
		t.Errorf("files %v, want the icon copy and no bundle file", f.files.data)
	}
	if !f.repo.bundles[0].Archived() {
		t.Error("the archive time was not saved")
	}
	if _, err := f.svc.Archive(context.Background(), ref); err != nil {
		t.Errorf("archiving twice: %v", err)
	}
}

func TestArchiveWithoutFileOrIcon(t *testing.T) {
	tests := []struct {
		name string
		b    models.AppBinary
	}{
		{"no file path", models.AppBinary{Info: &hubapp.BundleInfo{Icon: "i.png"}}},
		{"no info", models.AppBinary{AppPath: "a.apk"}},
		{"no icon", models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			got, err := f.svc.Archive(context.Background(), f.addBundle(tt.b))
			if err != nil || !got.Archived() || f.files.has("/store/a.apk.icon") {
				t.Fatalf("archive %+v, %v, files %v", got, err, f.files.data)
			}
		})
	}
}

func TestArchiveKeepsNoIconWhenItCannotBeRead(t *testing.T) {
	f := newFixture(t)
	ref := archivable(f)
	f.analyser.iconErr = hubapp.ErrNoIcon
	if _, err := f.svc.Archive(context.Background(), ref); err != nil || f.files.has("/store/a.apk.icon") {
		t.Fatalf("archive: %v, files %v", err, f.files.data)
	}
}

func TestArchiveRefusesABundleARunUses(t *testing.T) {
	f := newFixture(t)
	ref := archivable(f)
	f.repo.runs = []models.TestRun{{Model: models.Model{ID: 3}, AppBinaryID: ref.ID}}
	f.running[3] = true
	_, err := f.svc.Archive(context.Background(), ref)
	wantKind(t, err, apperr.KindConflict)
	if f.repo.bundles[0].Archived() || !f.files.has("/store/a.apk") {
		t.Error("a bundle in use must stay as it is")
	}
	f.running[3] = false
	if _, err := f.svc.Archive(context.Background(), ref); err != nil {
		t.Fatalf("archive after the run: %v", err)
	}
}

func TestArchiveErrors(t *testing.T) {
	f := newFixture(t)
	ref := archivable(f)
	_, err := f.svc.Archive(context.Background(), BundleRef{Scope: Scope{ProjectID: 2, AppID: 1}, ID: ref.ID})
	wantKind(t, err, apperr.KindNotFound)
	for _, method := range []string{"RecentRunIDs", "SaveArchivedAt"} {
		f.repo.fail[method] = errBoom
		if _, err := f.svc.Archive(context.Background(), ref); err == nil {
			t.Errorf("%s: a store error must come back", method)
		}
		delete(f.repo.fail, method)
	}
	f.files.fail["Remove:/store/a.apk"] = errBoom
	_, err = f.svc.Archive(context.Background(), ref)
	if err == nil || apperr.KindOf(err) != apperr.KindInternal || !f.repo.bundles[0].Archived() {
		t.Fatalf("a file that stays: %v, want an error and an archived bundle", err)
	}
}
