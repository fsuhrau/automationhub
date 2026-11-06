package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if err := db.AutoMigrate(&models.App{}, &models.AppBinary{}, &models.TestRun{}); err != nil {
		t.Fatal(err)
	}
	store := t.TempDir()
	old := apps.AppStoragePath
	apps.AppStoragePath = store
	t.Cleanup(func() { apps.AppStoragePath = old })

	app := models.App{Name: "game"}
	db.Create(&app)

	gin.SetMode(gin.TestMode)
	s := &Service{db: db}
	r := gin.New()
	r.GET("/bundles", func(c *gin.Context) { s.getBinaries(c, nil, &app) })
	r.PUT("/bundle/:binary_id", func(c *gin.Context) { s.updateBinary(c, nil, &app) })
	r.DELETE("/bundle/:binary_id", func(c *gin.Context) { s.deleteBinary(c, nil, &app) })
	return r, db, app, store
}

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
