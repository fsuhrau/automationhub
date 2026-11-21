package app

import (
	"context"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func pngBytes() []byte { return []byte("\x89PNG\r\n\x1a\nfake") }

func TestIconOfALiveBundleComesOutOfItsFile(t *testing.T) {
	f := newFixture(t)
	f.analyser.icon, f.analyser.iconType = []byte("data"), "image/png"
	ref := f.addBundle(models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{Icon: "i.png"}})
	got, err := f.svc.Icon(context.Background(), ref)
	if err != nil || string(got.Data) != "data" || got.ContentType != "image/png" {
		t.Fatalf("icon %+v, %v", got, err)
	}
}

func TestIconWithoutOne(t *testing.T) {
	tests := []struct {
		name string
		b    models.AppBinary
		err  error
	}{
		{"no info", models.AppBinary{AppPath: "a.apk"}, nil},
		{"no icon in the info", models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{}}, nil},
		{"no file path", models.AppBinary{Info: &hubapp.BundleInfo{Icon: "i.png"}}, nil},
		{"the file has none", models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{Icon: "i.png"}}, hubapp.ErrNoIcon},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.analyser.iconErr = tt.err
			_, err := f.svc.Icon(context.Background(), f.addBundle(tt.b))
			wantKind(t, err, apperr.KindNotFound)
			if err.Error() != hubapp.ErrNoIcon.Error() {
				t.Errorf("message %q, want %q", err, hubapp.ErrNoIcon)
			}
		})
	}
}

func TestIconErrors(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{AppPath: "a.apk", Info: &hubapp.BundleInfo{Icon: "i.png"}})
	f.analyser.iconErr = errBoom
	_, err := f.svc.Icon(context.Background(), ref)
	wantKind(t, err, apperr.KindInternal)
	_, err = f.svc.Icon(context.Background(), BundleRef{Scope: Scope{ProjectID: 2, AppID: 1}, ID: ref.ID})
	wantKind(t, err, apperr.KindNotFound)
}

func TestIconOfAnArchivedBundleIsTheKeptCopy(t *testing.T) {
	f := newFixture(t)
	f.files.data["/store/a.apk.icon"] = pngBytes()
	ref := f.addBundle(models.AppBinary{AppPath: "a.apk", ArchivedAt: archived()})
	got, err := f.svc.Icon(context.Background(), ref)
	if err != nil || got.ContentType != "image/png" || string(got.Data) != string(pngBytes()) {
		t.Fatalf("icon %+v, %v", got, err)
	}
}

func TestIconOfAnArchivedBundleWithoutACopy(t *testing.T) {
	f := newFixture(t)
	for _, path := range []string{"a.apk", ""} {
		_, err := f.svc.Icon(context.Background(), f.addBundle(models.AppBinary{AppPath: path, ArchivedAt: archived()}))
		wantKind(t, err, apperr.KindNotFound)
	}
	f.files.fail["Read"] = errBoom
	_, err := f.svc.Icon(context.Background(), f.addBundle(models.AppBinary{AppPath: "b.apk", ArchivedAt: archived()}))
	wantKind(t, err, apperr.KindInternal)
}

func readParams() hubapp.Parameter {
	return hubapp.Parameter{
		Platform: "ios", Identifier: "com.game", Name: "Game", Version: "45",
		Info: &hubapp.BundleInfo{Version: "1.2.3"},
		App: &hubapp.AppParams{
			Android:    &hubapp.AndroidParams{LaunchActivity: "Main"},
			Executable: &hubapp.ExecutableParams{Executable: "game"},
		},
	}
}

func TestReadAgainUpdatesWhatTheFileSays(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = readParams()
	ref := f.addBundle(models.AppBinary{AppPath: "a.ipa", Hash: "abc", Tags: "rc1"})
	got, err := f.svc.ReadAgain(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}
	if got.Identifier != "com.game" || got.Version != "45" || got.Platform != "ios" || got.Name != "Game" ||
		got.Android.LaunchActivity != "Main" || got.Executable.Executable != "game" {
		t.Errorf("read %+v", got.AppBinary)
	}
	if got.Hash != "abc" || got.Tags != "rc1" || got.Info.Version != "1.2.3" {
		t.Errorf("hash, tags and info %q %q %+v", got.Hash, got.Tags, got.Info)
	}
}

func TestReadAgainKeepsWhatTheFileDoesNotSay(t *testing.T) {
	f := newFixture(t)
	f.analyser.params = hubapp.Parameter{App: &hubapp.AppParams{}}
	ref := f.addBundle(models.AppBinary{AppPath: "a.ipa", Name: "Kept", Version: "9"})
	got, err := f.svc.ReadAgain(context.Background(), ref)
	if err != nil || got.Name != "Kept" || got.Version != "9" {
		t.Fatalf("read %+v, %v", got, err)
	}
}

func TestReadAgainErrors(t *testing.T) {
	f := newFixture(t)
	archived := f.addBundle(models.AppBinary{AppPath: "a.ipa", ArchivedAt: archived()})
	noFile := f.addBundle(models.AppBinary{})
	live := f.addBundle(models.AppBinary{AppPath: "b.ipa"})
	ctx := context.Background()

	_, err := f.svc.ReadAgain(ctx, archived)
	wantKind(t, err, apperr.KindConflict)
	_, err = f.svc.ReadAgain(ctx, noFile)
	wantKind(t, err, apperr.KindConflict)
	_, err = f.svc.ReadAgain(ctx, BundleRef{Scope: f.sc, ID: 99})
	wantKind(t, err, apperr.KindNotFound)

	f.analyser.readErr = errBoom
	_, err = f.svc.ReadAgain(ctx, live)
	wantKind(t, err, apperr.KindInvalid)

	f.analyser.readErr = nil
	f.analyser.params = readParams()
	f.repo.fail["SaveRead"] = errBoom
	if _, err := f.svc.ReadAgain(ctx, live); err == nil {
		t.Fatal("a store error must come back")
	}
}
