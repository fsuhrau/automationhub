package base

import (
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
)

func TestGetParamsUsesTheBundlePackage(t *testing.T) {
	registered := &models.App{Identifier: "com.innogames.heroesofhistory"}

	feature := &models.AppBinary{App: registered, Identifier: "com.innogames.enterprise.heroesofhistory.feature"}
	if got := GetParams(feature, "").Identifier; got != "com.innogames.enterprise.heroesofhistory.feature" {
		t.Fatalf("feature build: %q", got)
	}

	// bundles without an analysed package keep the app's
	old := &models.AppBinary{App: registered}
	if got := GetParams(old, "").Identifier; got != "com.innogames.heroesofhistory" {
		t.Fatalf("fallback: %q", got)
	}

	if got := GetParams(&models.AppBinary{}, "").Identifier; got != "" {
		t.Fatalf("no app loaded: %q", got)
	}
}
