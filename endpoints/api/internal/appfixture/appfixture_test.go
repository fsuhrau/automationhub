package appfixture_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/appfixture"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestEnvRestoresTheStorePath(t *testing.T) {
	old := apps.AppStoragePath
	t.Run("env", func(t *testing.T) { checkEnv(t) })
	if apps.AppStoragePath != old {
		t.Errorf("the store path is %q, want %q back", apps.AppStoragePath, old)
	}
}

func checkEnv(t *testing.T) {
	e := appfixture.New(t)
	if apps.AppStoragePath != e.Store || e.Mine.ProjectID == e.Other.ProjectID {
		t.Fatalf("store %q (%q), projects %d and %d", apps.AppStoragePath, e.Store, e.Mine.ProjectID, e.Other.ProjectID)
	}
	b := e.AddBundle(models.AppBinary{Version: "1"})
	if b.ID == 0 || b.AppID != e.Mine.ID {
		t.Fatalf("bundle %+v", b)
	}
	if rec := e.Do("GET", e.BundlePath(b.ID, ""), nil); rec.Code != http.StatusOK {
		t.Errorf("get: %d %s", rec.Code, rec.Body)
	}
	if rec := e.Raw("PUT", e.BundlePath(b.ID, ""), "{"); rec.Code != http.StatusBadRequest {
		t.Errorf("raw: %d", rec.Code)
	}
	if _, err := os.Stat(e.WriteIPA("a.ipa")); err != nil {
		t.Error(err)
	}
	if rec := e.Upload("test_target", "Game.ipa", appfixture.IPA(t)); rec.Code != http.StatusCreated {
		t.Errorf("upload: %d %s", rec.Code, rec.Body)
	}
}
