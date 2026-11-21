package bundles_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/endpoints/api/internal/appfixture"
	"github.com/fsuhrau/automationhub/internal/testkit"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

func newBundleEnv(t *testing.T) *appfixture.Env { return appfixture.New(t) }

func TestBundleListIsNewestFirstWithRunCounts(t *testing.T) {
	e := newBundleEnv(t)
	older := e.AddBundle(models.AppBinary{Version: "1", Info: &hubapp.BundleInfo{Identifier: "com.example", Android: &hubapp.AndroidInfo{MinSdk: "24"}}})
	e.AddBundle(models.AppBinary{Version: "2"})
	e.DB.Create(&models.AppBinary{AppID: e.Other.ID, Version: "other app"})
	for i := 0; i < 3; i++ {
		e.DB.Create(&models.TestRun{AppBinaryID: older.ID})
	}

	rec := e.Do("GET", e.AppPath("/bundles"), nil)
	list := testkit.Decode[[]appsvc.BundleEntry](t, rec)
	if len(list) != 2 || list[0].Version != "2" || list[0].RunCount != 0 || list[1].Version != "1" || list[1].RunCount != 3 {
		t.Fatalf("list %+v, want version 2 with 0 runs, then version 1 with 3", list)
	}
	if !strings.Contains(rec.Body.String(), `"identifier":"com.example"`) || strings.Contains(rec.Body.String(), `"minSdk"`) {
		t.Errorf("list info: %s", rec.Body)
	}
}

func TestBundleOfAnotherProjectsAppIsNotFound(t *testing.T) {
	e := newBundleEnv(t)
	foreign := models.AppBinary{AppID: e.Other.ID, Version: "x"}
	e.DB.Create(&foreign)
	mine := e.AddBundle(models.AppBinary{Version: "1"})

	for _, route := range []struct{ method, path string }{
		{"GET", fmt.Sprintf("/api/mine/app/%d/bundles", e.Other.ID)},
		{"GET", e.BundlePath(foreign.ID, "")},
		{"GET", e.BundlePath(foreign.ID, "/icon")},
		{"POST", e.BundlePath(foreign.ID, "/read")},
		{"POST", e.BundlePath(foreign.ID, "/archive")},
		{"PUT", e.BundlePath(foreign.ID, "")},
		{"DELETE", e.BundlePath(foreign.ID, "")},
		{"GET", fmt.Sprintf("/api/other/app/%d/bundle/%d", e.Other.ID, mine.ID)},
	} {
		if rec := e.Do(route.method, route.path, map[string]string{"tags": "x"}); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s: %d, want 404", route.method, route.path, rec.Code)
		}
	}
	var n int64
	e.DB.Model(&models.AppBinary{}).Count(&n)
	if n != 2 {
		t.Errorf("%d bundles, want both: nothing may be deleted or inserted", n)
	}
}

func TestBundleNonNumericIDIsBadRequest(t *testing.T) {
	e := newBundleEnv(t)
	for _, route := range []struct{ method, suffix string }{
		{"GET", ""}, {"GET", "/icon"}, {"POST", "/read"}, {"POST", "/archive"}, {"PUT", ""}, {"DELETE", ""},
	} {
		if rec := e.Do(route.method, e.AppPath("/bundle/abc"+route.suffix), nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d, want 400", route.method, route.suffix, rec.Code)
		}
	}
}

func TestBundleTags(t *testing.T) {
	e := newBundleEnv(t)
	bundle := e.AddBundle(models.AppBinary{Version: "1"})
	rec := e.Do("PUT", e.BundlePath(bundle.ID, ""), map[string]string{"tags": " nightly, ,rc1,nightly "})
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	e.DB.First(bundle, bundle.ID)
	if bundle.Tags != "nightly,rc1" {
		t.Errorf("tags %q", bundle.Tags)
	}
	if rec := e.Do("PUT", e.BundlePath(999, ""), map[string]string{"tags": "x"}); rec.Code != http.StatusNotFound {
		t.Errorf("missing bundle: %d", rec.Code)
	}
	req := httptest.NewRequest("PUT", e.BundlePath(bundle.ID, ""), strings.NewReader("tags=x"))
	if rec := e.Request(req); rec.Code != http.StatusBadRequest {
		t.Errorf("not JSON: %d", rec.Code)
	}
	var count int64
	e.DB.Model(&models.AppBinary{}).Count(&count)
	if count != 1 {
		t.Errorf("%d bundles, want the one; an update must not insert", count)
	}
}

// Deleting a bundle removes its file and nothing else: a bundle without a
// path used to remove the whole storage folder.
func TestBundleDelete(t *testing.T) {
	e := newBundleEnv(t)
	write := func(path string) string {
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	mine := write(filepath.Join(e.Store, "game_abc.apk"))
	others := write(filepath.Join(e.Store, "other_def.apk"))
	outside := write(filepath.Join(filepath.Dir(e.Store), "outside.txt"))
	t.Cleanup(func() { _ = os.Remove(outside) })

	bundle := e.AddBundle(models.AppBinary{AppPath: "game_abc.apk"})
	empty := e.AddBundle(models.AppBinary{AppPath: ""})
	escape := e.AddBundle(models.AppBinary{AppPath: "../outside.txt"})

	for _, b := range []*models.AppBinary{empty, escape, bundle} {
		if rec := e.Do("DELETE", e.BundlePath(b.ID, ""), nil); rec.Code != http.StatusOK {
			t.Fatalf("delete %q: %d %s", b.AppPath, rec.Code, rec.Body)
		}
	}
	for path, want := range map[string]bool{mine: false, others: true, outside: true} {
		if _, err := os.Stat(path); (err == nil) != want {
			t.Errorf("%s exists %v, want %v", path, err == nil, want)
		}
	}
	if rec := e.Do("DELETE", e.BundlePath(999, ""), nil); rec.Code != http.StatusNotFound {
		t.Errorf("missing bundle: %d", rec.Code)
	}
}

func TestBundleDetailListsTheRunsThatUsedIt(t *testing.T) {
	e := newBundleEnv(t)
	test := testkit.Test(e.DB, func(x *models.Test) { x.AppID, x.Name = e.Mine.ID, "smoke" })
	bundle := e.AddBundle(models.AppBinary{Version: "1", Info: &hubapp.BundleInfo{
		Format: "apk", Identifier: "com.example", Android: &hubapp.AndroidInfo{MinSdk: "24"},
	}})
	e.DB.Create(&models.TestRun{AppBinaryID: bundle.ID, TestID: test.ID})

	rec := e.Do("GET", e.BundlePath(bundle.ID, ""), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	detail := testkit.Decode[appsvc.BundleDetail](t, rec)
	if detail.Info == nil || detail.Info.Android == nil || detail.Info.Android.MinSdk != "24" {
		t.Errorf("info %+v", detail.Info)
	}
	if detail.RunCount != 1 || len(detail.Runs) != 1 || detail.Runs[0].TestName != "smoke" || detail.Runs[0].Status != "empty" {
		t.Errorf("runs %d %+v", detail.RunCount, detail.Runs)
	}
}

// A bundle uploaded before the master read everything is read again from
// its file; the hash and tags stay.
func TestBundleReadAgain(t *testing.T) {
	e := newBundleEnv(t)
	e.WriteIPA("game_abc.ipa")
	bundle := e.AddBundle(models.AppBinary{AppPath: "game_abc.ipa", Hash: "abc", Tags: "rc1"})

	if rec := e.Do("GET", e.BundlePath(bundle.ID, "/icon"), nil); rec.Code != http.StatusNotFound {
		t.Errorf("icon before reading: %d", rec.Code)
	}
	if rec := e.Do("POST", e.BundlePath(bundle.ID, "/read"), nil); rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var stored models.AppBinary
	e.DB.First(&stored, bundle.ID)
	if stored.Identifier != "com.example.game" || stored.Version != "45" || stored.Platform != "ios" || stored.Name != "Game" {
		t.Errorf("stored %+v", stored)
	}
	if stored.Hash != "abc" || stored.Tags != "rc1" || stored.Info == nil || stored.Info.Version != "1.2.3" {
		t.Errorf("hash %q tags %q info %+v", stored.Hash, stored.Tags, stored.Info)
	}
	rec := e.Do("GET", e.BundlePath(bundle.ID, "/icon"), nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("Cache-Control") == "" {
		t.Errorf("icon: %d %s", rec.Code, rec.Header())
	}

	noFile := e.AddBundle(models.AppBinary{})
	if rec := e.Do("POST", e.BundlePath(noFile.ID, "/read"), nil); rec.Code != http.StatusConflict {
		t.Errorf("bundle without file: %d", rec.Code)
	}
}

func TestBundleReadAgainOfAFileThatIsNoBundle(t *testing.T) {
	e := newBundleEnv(t)
	if err := os.WriteFile(filepath.Join(e.Store, "junk.ipa"), []byte("junk"), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle := e.AddBundle(models.AppBinary{AppPath: "junk.ipa"})
	if rec := e.Do("POST", e.BundlePath(bundle.ID, "/read"), nil); rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "unable to read the bundle") {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestUploadStoresTheBundle(t *testing.T) {
	e := newBundleEnv(t)
	rec := e.Upload("test_target", "Game.ipa", appfixture.IPA(t))
	if rec.Code != http.StatusCreated {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	got := testkit.Decode[models.AppBinary](t, rec)
	if got.AppID != e.Mine.ID || got.Identifier != "com.example.game" || !strings.HasPrefix(got.AppPath, "Game_") {
		t.Errorf("bundle %+v", got)
	}
	if _, err := os.Stat(filepath.Join(e.Store, got.AppPath)); err != nil {
		t.Errorf("file not stored: %v", err)
	}
	if rec := e.Upload("test_target", "Game.ipa", appfixture.IPA(t)); rec.Code != http.StatusConflict {
		t.Errorf("same build again: %d %s", rec.Code, rec.Body)
	}
}

func TestUploadErrors(t *testing.T) {
	e := newBundleEnv(t)
	if rec := e.Upload("other_field", "Game.ipa", appfixture.IPA(t)); rec.Code != http.StatusBadRequest {
		t.Errorf("wrong field: %d", rec.Code)
	}
	rec := e.Upload("test_target", "junk.ipa", []byte("junk"))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "unable to analyse app") {
		t.Errorf("a file no analyser reads: %d %s", rec.Code, rec.Body)
	}
	if _, err := os.Stat(filepath.Join(e.Store, "junk.ipa")); !os.IsNotExist(err) {
		t.Error("the unreadable upload stays")
	}
	if rec := e.Do("POST", fmt.Sprintf("/api/mine/app/%d/upload", e.Other.ID), nil); rec.Code != http.StatusNotFound {
		t.Errorf("another project's app: %d", rec.Code)
	}
}

func TestBundleListFailsWhenTheRunsCannotBeCounted(t *testing.T) {
	e := newBundleEnv(t)
	e.AddBundle(models.AppBinary{Version: "1"})
	if err := e.DB.Migrator().DropTable(&models.TestRun{}); err != nil {
		t.Fatal(err)
	}
	if rec := e.Do("GET", e.AppPath("/bundles"), nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}
