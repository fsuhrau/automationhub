package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func bundlesSetup(t *testing.T) (*gin.Engine, *gorm.DB, models.App, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.App{}, &models.AppBinary{}, &models.TestRun{}, &models.Test{}, &models.TestProtocol{}, &models.TestRunLogEntry{}); err != nil {
		t.Fatal(err)
	}
	store := t.TempDir()
	old := apps.AppStoragePath
	apps.AppStoragePath = store
	t.Cleanup(func() { apps.AppStoragePath = old })

	app := models.App{Name: "game"}
	db.Create(&app)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db, runTracker: runs.NewTracker(nil)}
	testService = s
	r := gin.New()
	r.POST("/bundle/:binary_id/archive", func(c *gin.Context) { s.archiveBinary(c, nil, &app) })
	r.GET("/bundle/:binary_id", func(c *gin.Context) { s.getBinary(c, nil, &app) })
	r.GET("/bundle/:binary_id/icon", func(c *gin.Context) { s.getBinaryIcon(c, nil, &app) })
	r.POST("/bundle/:binary_id/read", func(c *gin.Context) { s.readBinaryAgain(c, nil, &app) })
	r.GET("/bundles", func(c *gin.Context) { s.getBinaries(c, nil, &app) })
	r.PUT("/bundle/:binary_id", func(c *gin.Context) { s.updateBinary(c, nil, &app) })
	r.DELETE("/bundle/:binary_id", func(c *gin.Context) { s.deleteBinary(c, nil, &app) })
	return r, db, app, store
}

// testService is the service of the last bundlesSetup
var testService *Service

func send(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// The bundles page lists the newest first, with how many runs used each.
func TestBundleList(t *testing.T) {
	r, db, app, _ := bundlesSetup(t)
	older := models.AppBinary{AppID: app.ID, Version: "1"}
	newer := models.AppBinary{AppID: app.ID, Version: "2"}
	db.Create(&older)
	db.Create(&newer)
	db.Create(&models.AppBinary{AppID: app.ID + 1, Version: "other app"})
	for i := 0; i < 3; i++ {
		db.Create(&models.TestRun{AppBinaryID: older.ID})
	}

	w := send(r, "GET", "/bundles", "")
	var list []bundleListEntry
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("%v: %s", err, w.Body)
	}
	got := fmt.Sprint(len(list), " ", list[0].Version, list[0].RunCount, " ", list[1].Version, list[1].RunCount)
	if got != "2 20 13" {
		t.Errorf("list %s, want 2 bundles: version 2 with 0 runs, version 1 with 3", got)
	}
	if !strings.Contains(w.Body.String(), `"runCount":3`) || !strings.Contains(w.Body.String(), `"version":"1"`) {
		t.Errorf("bundle fields not at the top level: %s", w.Body)
	}
}

// Tags are tidied, and a missing bundle is a 404. Before, a missing bundle
// made the update insert a new empty bundle.
func TestBundleTags(t *testing.T) {
	r, db, app, _ := bundlesSetup(t)
	bundle := models.AppBinary{AppID: app.ID, Version: "1"}
	db.Create(&bundle)

	w := send(r, "PUT", fmt.Sprintf("/bundle/%d", bundle.ID), `{"tags":" nightly, ,rc1,nightly "}`)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	db.First(&bundle, bundle.ID)
	if bundle.Tags != "nightly,rc1" {
		t.Errorf("tags %q", bundle.Tags)
	}

	if w := send(r, "PUT", "/bundle/999", `{"tags":"x"}`); w.Code != http.StatusNotFound {
		t.Errorf("missing bundle: %d", w.Code)
	}
	if w := send(r, "PUT", fmt.Sprintf("/bundle/%d", bundle.ID), `tags=x`); w.Code != http.StatusBadRequest {
		t.Errorf("not JSON: %d", w.Code)
	}
	var count int64
	db.Model(&models.AppBinary{}).Count(&count)
	if count != 1 {
		t.Errorf("%d bundles, want the one; an update must not insert", count)
	}
}

// Deleting a bundle removes its file and nothing else: a bundle without a
// path used to remove the whole storage folder.
func TestBundleDelete(t *testing.T) {
	r, db, app, store := bundlesSetup(t)
	write := func(name string) string {
		path := filepath.Join(store, name)
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	mine := write("game_abc.apk")
	others := write("other_def.apk")
	outside := filepath.Join(filepath.Dir(store), "outside.txt")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(outside) })

	bundle := models.AppBinary{AppID: app.ID, AppPath: "game_abc.apk"}
	empty := models.AppBinary{AppID: app.ID, AppPath: ""}
	escape := models.AppBinary{AppID: app.ID, AppPath: "../outside.txt"}
	db.Create(&bundle)
	db.Create(&empty)
	db.Create(&escape)

	for _, b := range []models.AppBinary{empty, escape, bundle} {
		if w := send(r, "DELETE", fmt.Sprintf("/bundle/%d", b.ID), ""); w.Code != http.StatusOK {
			t.Fatalf("delete %q: %d %s", b.AppPath, w.Code, w.Body)
		}
	}
	if _, err := os.Stat(mine); !os.IsNotExist(err) {
		t.Errorf("the bundle's file is still there")
	}
	if _, err := os.Stat(others); err != nil {
		t.Errorf("another bundle's file is gone: %v", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("a file outside the store is gone: %v", err)
	}
	if w := send(r, "DELETE", "/bundle/999", ""); w.Code != http.StatusNotFound {
		t.Errorf("missing bundle: %d", w.Code)
	}
}

// writeIPA stores a minimal ipa in the bundle storage
func writeIPA(t *testing.T, store, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join(store, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	plist, _ := w.Create("Payload/Game.app/Info.plist")
	plist.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.example.game</string>
<key>CFBundleShortVersionString</key><string>1.2.3</string>
<key>CFBundleVersion</key><string>45</string>
<key>CFBundleDisplayName</key><string>Game</string>
<key>MinimumOSVersion</key><string>15.0</string>
<key>CFBundleIconFiles</key><array><string>Icon</string></array>
</dict></plist>`))
	icon, _ := w.Create("Payload/Game.app/Icon@2x.png")
	icon.Write([]byte("\x89PNG\r\n\x1a\nnot crushed"))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// The bundle page gets the bundle with its info and the runs that used it;
// the list only gets a summary of the info.
func TestBundleDetail(t *testing.T) {
	r, db, app, _ := bundlesSetup(t)
	test := models.Test{AppID: app.ID, Name: "smoke"}
	db.Create(&test)
	bundle := models.AppBinary{AppID: app.ID, Version: "1", Info: &hubapp.BundleInfo{
		Format: "apk", Identifier: "com.example", Android: &hubapp.AndroidInfo{MinSdk: "24"},
	}}
	db.Create(&bundle)
	db.Create(&models.TestRun{AppBinaryID: bundle.ID, TestID: test.ID})

	w := send(r, "GET", fmt.Sprintf("/bundle/%d", bundle.ID), "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var detail bundleDetail
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Info == nil || detail.Info.Android == nil || detail.Info.Android.MinSdk != "24" {
		t.Errorf("info %+v", detail.Info)
	}
	if detail.RunCount != 1 || len(detail.Runs) != 1 || detail.Runs[0].TestName != "smoke" || detail.Runs[0].Status != "empty" {
		t.Errorf("runs %d %+v", detail.RunCount, detail.Runs)
	}
	if w := send(r, "GET", "/bundle/999", ""); w.Code != http.StatusNotFound {
		t.Errorf("missing bundle: %d", w.Code)
	}

	w = send(r, "GET", "/bundles", "")
	if !strings.Contains(w.Body.String(), `"identifier":"com.example"`) || strings.Contains(w.Body.String(), `"minSdk"`) {
		t.Errorf("list info: %s", w.Body)
	}
}

// A bundle uploaded before the master read everything is read again from
// its file; the hash and tags stay.
func TestBundleReadAgain(t *testing.T) {
	r, db, app, store := bundlesSetup(t)
	writeIPA(t, store, "game_abc.ipa")
	bundle := models.AppBinary{AppID: app.ID, AppPath: "game_abc.ipa", Hash: "abc", Tags: "rc1"}
	db.Create(&bundle)

	if w := send(r, "GET", fmt.Sprintf("/bundle/%d/icon", bundle.ID), ""); w.Code != http.StatusNotFound {
		t.Errorf("icon before reading: %d", w.Code)
	}
	w := send(r, "POST", fmt.Sprintf("/bundle/%d/read", bundle.ID), "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var stored models.AppBinary
	db.First(&stored, bundle.ID)
	if stored.Identifier != "com.example.game" || stored.Version != "45" || stored.Platform != "ios" || stored.Name != "Game" {
		t.Errorf("stored %+v", stored)
	}
	if stored.Hash != "abc" || stored.Tags != "rc1" || stored.Info == nil || stored.Info.Version != "1.2.3" {
		t.Errorf("hash %q tags %q info %+v", stored.Hash, stored.Tags, stored.Info)
	}
	w = send(r, "GET", fmt.Sprintf("/bundle/%d/icon", bundle.ID), "")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("icon: %d %s", w.Code, w.Header().Get("Content-Type"))
	}

	missing := models.AppBinary{AppID: app.ID, AppPath: ""}
	db.Create(&missing)
	if w := send(r, "POST", fmt.Sprintf("/bundle/%d/read", missing.ID), ""); w.Code != http.StatusConflict {
		t.Errorf("bundle without file: %d", w.Code)
	}
}

// Archiving removes the file but keeps the bundle, its info and its icon;
// it can't be read again, and uploading the same file brings it back.
func TestBundleArchive(t *testing.T) {
	r, db, app, store := bundlesSetup(t)
	writeIPA(t, store, "upload.ipa")
	bundle, err := testService.StoreBundle(&app, filepath.Join(store, "upload.ipa"))
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(store, bundle.AppPath)

	running := models.TestRun{AppBinaryID: bundle.ID}
	db.Create(&running)
	testService.runTracker.Started(running.ID)
	if w := send(r, "POST", fmt.Sprintf("/bundle/%d/archive", bundle.ID), ""); w.Code != http.StatusConflict {
		t.Errorf("archive while a run uses it: %d", w.Code)
	}
	testService.runTracker.Finished(running.ID, runs.RunResult{})

	w := send(r, "POST", fmt.Sprintf("/bundle/%d/archive", bundle.ID), "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), `"archivedAt"`) || !strings.Contains(w.Body.String(), `"runCount":1`) {
		t.Errorf("archived detail: %s", w.Body)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("file stays: %v", err)
	}
	var stored models.AppBinary
	db.First(&stored, bundle.ID)
	if !stored.Archived() || stored.Info == nil || stored.Info.Identifier != "com.example.game" || stored.Hash != bundle.Hash {
		t.Errorf("stored %+v", stored)
	}
	if w := send(r, "GET", fmt.Sprintf("/bundle/%d/icon", bundle.ID), ""); w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("icon of archived bundle: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	if w := send(r, "POST", fmt.Sprintf("/bundle/%d/read", bundle.ID), ""); w.Code != http.StatusConflict {
		t.Errorf("read archived bundle: %d", w.Code)
	}
	if w := send(r, "POST", fmt.Sprintf("/bundle/%d/archive", bundle.ID), ""); w.Code != http.StatusOK {
		t.Errorf("archive twice: %d", w.Code)
	}

	writeIPA(t, store, "again.ipa")
	restored, err := testService.StoreBundle(&app, filepath.Join(store, "again.ipa"))
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != bundle.ID || restored.Archived() {
		t.Errorf("restored %d archived %v, want bundle %d back", restored.ID, restored.Archived(), bundle.ID)
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("file not back: %v", err)
	}
	if _, err := os.Stat(file + ".icon"); !os.IsNotExist(err) {
		t.Errorf("icon copy stays: %v", err)
	}
	var count int64
	db.Model(&models.AppBinary{}).Count(&count)
	if count != 1 {
		t.Errorf("%d bundles, want 1", count)
	}
}
