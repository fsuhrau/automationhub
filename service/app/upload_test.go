package app

import (
	"context"
	"strings"
	"testing"
	"time"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// archived is an archive time for bundles that are archived.
func archived() *time.Time {
	t := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return &t
}

func uploadParams() hubapp.Parameter {
	return hubapp.Parameter{
		Platform: "android", Identifier: "com.game", Name: "Game", Version: "1.0",
		Info: &hubapp.BundleInfo{Format: "apk"},
		App: &hubapp.AppParams{
			Hash: "h1", Size: 12, Additional: "extra",
			Android:    &hubapp.AndroidParams{LaunchActivity: "Main"},
			Executable: &hubapp.ExecutableParams{Executable: "game"},
		},
	}
}

func TestUploadStoresTheBundle(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = uploadParams()
	got, err := f.svc.Upload(context.Background(), f.sc, UploadFile{Name: "dir/Game.apk", Content: strings.NewReader("apk")})
	if err != nil {
		t.Fatal(err)
	}
	if got.AppPath != "Game_h1.apk" || got.Hash != "h1" || got.AppID != 1 || got.Name != "Game" || got.Identifier != "com.game" ||
		got.Platform != "android" || got.Size != 12 || got.Additional != "extra" || got.Info.Format != "apk" ||
		got.Android.LaunchActivity != "Main" || got.Executable.Executable != "game" {
		t.Errorf("bundle %+v", got)
	}
	if !f.files.has("/store/Game_h1.apk") || f.files.has("/store/dir/Game.apk") || string(f.files.data["/store/Game_h1.apk"]) != "apk" {
		t.Errorf("files %v, want the upload renamed to name_hash.ext", f.files.data)
	}
	if len(f.repo.bundles) != 1 || len(f.analyser.analysed) != 1 {
		t.Errorf("%d bundles, %d analysed", len(f.repo.bundles), len(f.analyser.analysed))
	}
}

func TestUploadWithoutPlatformDetails(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = hubapp.Parameter{App: &hubapp.AppParams{Hash: "h2"}}
	got, err := f.svc.StoreBundle(context.Background(), f.sc, "/store/x")
	if err != nil || got.Android.LaunchActivity != "" || got.Executable.Executable != "" || got.AppPath != "x_h2" {
		t.Fatalf("bundle %+v, %v", got, err)
	}
}

func TestUploadAnalyserErrorRemovesTheUpload(t *testing.T) {
	f := newFixture(t)
	f.analyser.analyseErr = errBoom
	_, err := f.svc.Upload(context.Background(), f.sc, UploadFile{Name: "Game.apk", Content: strings.NewReader("junk")})
	if err == nil || !strings.Contains(err.Error(), "unable to analyse app") {
		t.Fatalf("error %v", err)
	}
	if len(f.files.data) != 0 || len(f.repo.bundles) != 0 {
		t.Errorf("files %v, bundles %+v, want nothing left", f.files.data, f.repo.bundles)
	}
}

func TestUploadOfAKnownBuildKeepsTheStoredBundle(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = uploadParams()
	existing := f.addBundle(models.AppBinary{Hash: "h1", Version: "old"})
	got, err := f.svc.Upload(context.Background(), f.sc, UploadFile{Name: "Game.apk", Content: strings.NewReader("apk")})
	if err != ErrBundleExists || got == nil || got.ID != existing.ID {
		t.Fatalf("got %+v, %v; want the stored bundle and ErrBundleExists", got, err)
	}
	wantKind(t, err, apperr.KindConflict)
	if len(f.files.data) != 0 || len(f.repo.bundles) != 1 {
		t.Errorf("files %v, %d bundles; the duplicate must go", f.files.data, len(f.repo.bundles))
	}
}

func TestUploadOfAnArchivedBuildRestoresIt(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = uploadParams()
	existing := f.addBundle(models.AppBinary{Hash: "h1", AppPath: "Game_h1.apk", ArchivedAt: archived(), Tags: "rc"})
	f.files.data["/store/Game_h1.apk.icon"] = []byte("i")
	got, err := f.svc.Upload(context.Background(), f.sc, UploadFile{Name: "again.apk", Content: strings.NewReader("apk")})
	if err != nil || got.ID != existing.ID || got.Archived() || got.Tags != "rc" {
		t.Fatalf("got %+v, %v; want the bundle back", got, err)
	}
	if string(f.files.data["/store/Game_h1.apk"]) != "apk" || f.files.has("/store/Game_h1.apk.icon") {
		t.Errorf("files %v, want the file back and no icon copy", f.files.data)
	}
	if f.repo.bundles[0].Archived() {
		t.Error("the restore was not saved")
	}
}

func TestRestoreErrors(t *testing.T) {
	tests := []struct {
		name   string
		bundle models.AppBinary
		fail   map[string]error
		repo   string
	}{
		{"no file path", models.AppBinary{Hash: "h1", ArchivedAt: archived()}, nil, ""},
		{"file cannot move", models.AppBinary{Hash: "h1", AppPath: "a.apk", ArchivedAt: archived()}, map[string]error{"Move": errBoom}, ""},
		{"store error", models.AppBinary{Hash: "h1", AppPath: "a.apk", ArchivedAt: archived()}, nil, "SaveArchivedAt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.analyser.params = uploadParams()
			f.addBundle(tt.bundle)
			for k, v := range tt.fail {
				f.files.fail[k] = v
			}
			if tt.repo != "" {
				f.repo.fail[tt.repo] = errBoom
			}
			if _, err := f.svc.StoreBundle(context.Background(), f.sc, "/store/up.apk"); err == nil {
				t.Fatal("want an error")
			}
		})
	}
}

func TestUploadErrors(t *testing.T) {
	tests := []struct {
		name string
		sc   Scope
		file string
		repo string
	}{
		{"another project's app", Scope{ProjectID: 2, AppID: 1}, "", ""},
		{"file cannot be saved", Scope{ProjectID: 1, AppID: 1}, "Save", ""},
		{"file cannot move", Scope{ProjectID: 1, AppID: 1}, "Move", ""},
		{"hash lookup fails", Scope{ProjectID: 1, AppID: 1}, "", "BundleByHash"},
		{"bundle cannot be created", Scope{ProjectID: 1, AppID: 1}, "", "CreateBundle"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.analyser.params = uploadParams()
			if tt.file != "" {
				f.files.fail[tt.file] = errBoom
			}
			if tt.repo != "" {
				f.repo.fail[tt.repo] = errBoom
			}
			if _, err := f.svc.Upload(context.Background(), tt.sc, UploadFile{Name: "Game.apk", Content: strings.NewReader("apk")}); err == nil {
				t.Fatal("want an error")
			}
			if len(f.repo.bundles) != 0 {
				t.Errorf("bundles %+v, want none", f.repo.bundles)
			}
		})
	}
}

func TestStoreBundleOfAnotherProjectsAppIsNotFound(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = uploadParams()
	_, err := f.svc.StoreBundle(context.Background(), Scope{ProjectID: 2, AppID: 1}, "/store/x.apk")
	wantKind(t, err, apperr.KindNotFound)
	if len(f.analyser.analysed) != 0 {
		t.Error("the file of another project's app must not even be analysed")
	}
}
