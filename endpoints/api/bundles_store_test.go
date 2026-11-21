package api_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/endpoints/api/internal/appfixture"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

// Archiving removes the file but keeps the bundle, its info and its icon;
// it can't be read again, and uploading the same file brings it back.
func TestBundleArchiveAndRestore(t *testing.T) {
	e := appfixture.New(t)
	svc := api.New(logrus.New(), e.DB, "", e.Devices, nil, config.Service{}, e.Nodes)
	path := e.WriteIPA("upload.ipa")
	bundle, err := svc.StoreBundle(e.Mine, path)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(e.Store, bundle.AppPath)
	e.DB.Create(&models.TestRun{AppBinaryID: bundle.ID})

	rec := e.Do("POST", e.BundlePath(bundle.ID, "/archive"), nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"archivedAt"`) || !strings.Contains(rec.Body.String(), `"runCount":1`) {
		t.Fatalf("archive: %d %s", rec.Code, rec.Body)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("file stays: %v", err)
	}
	var stored models.AppBinary
	e.DB.First(&stored, bundle.ID)
	if !stored.Archived() || stored.Info == nil || stored.Info.Identifier != "com.example.game" || stored.Hash != bundle.Hash {
		t.Errorf("stored %+v", stored)
	}
	if rec := e.Do("GET", e.BundlePath(bundle.ID, "/icon"), nil); rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("icon of archived bundle: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec := e.Do("POST", e.BundlePath(bundle.ID, "/read"), nil); rec.Code != http.StatusConflict {
		t.Errorf("read archived bundle: %d", rec.Code)
	}
	if rec := e.Do("POST", e.BundlePath(bundle.ID, "/archive"), nil); rec.Code != http.StatusOK {
		t.Errorf("archive twice: %d", rec.Code)
	}

	restored, err := svc.StoreBundle(e.Mine, e.WriteIPA("again.ipa"))
	if err != nil || restored.ID != bundle.ID || restored.Archived() {
		t.Fatalf("restored %+v, %v, want bundle %d back", restored, err, bundle.ID)
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("file not back: %v", err)
	}
	if _, err := os.Stat(file + ".icon"); !os.IsNotExist(err) {
		t.Errorf("icon copy stays: %v", err)
	}
}

func TestStoreBundleReportsAnExistingBuildAndTheProjectScope(t *testing.T) {
	e := appfixture.New(t)
	svc := api.New(logrus.New(), e.DB, "", e.Devices, nil, config.Service{}, e.Nodes)
	first, err := svc.StoreBundle(e.Mine, e.WriteIPA("one.ipa"))
	if err != nil {
		t.Fatal(err)
	}
	again, err := svc.StoreBundle(e.Mine, e.WriteIPA("two.ipa"))
	if !errors.Is(err, api.ErrBundleExists) || again == nil || again.ID != first.ID {
		t.Fatalf("again %+v, %v; want the stored bundle and ErrBundleExists", again, err)
	}
	if _, err := os.Stat(filepath.Join(e.Store, "two.ipa")); !os.IsNotExist(err) {
		t.Error("the duplicate upload stays")
	}
	foreignApp := *e.Other
	foreignApp.ProjectID = e.Mine.ProjectID
	if _, err := svc.StoreBundle(&foreignApp, e.WriteIPA("three.ipa")); err == nil {
		t.Error("an app that is not the project's must not take a bundle")
	}
}
