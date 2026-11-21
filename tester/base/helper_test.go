package base

import (
	"strings"
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

func TestGetParams(t *testing.T) {
	binary := &models.AppBinary{
		Model: models.Model{ID: 3}, Name: "game", Version: "1.2", Identifier: "com.game", AppPath: "a/game.apk", Hash: "h", Size: 10,
		Android:    models.Android{LaunchActivity: ".Main"},
		Executable: models.Executable{Executable: "game.exe"},
	}
	got := GetParams(binary, "ignored")
	if got.Identifier != "com.game" || got.Name != "game" || got.Version != "1.2" || got.Web != nil {
		t.Fatalf("params %+v", got)
	}
	if got.App.AppBinaryID != 3 || got.App.Hash != "h" || got.App.Size != 10 || !strings.HasSuffix(got.App.AppPath, "a/game.apk") {
		t.Fatalf("app params %+v", got.App)
	}
	if got.App.Android.LaunchActivity != ".Main" || got.App.Executable.Executable != "game.exe" {
		t.Fatalf("platform params %+v", got.App)
	}

	plain := GetParams(&models.AppBinary{}, "")
	if plain.App.Android != nil || plain.App.Executable != nil {
		t.Fatalf("empty platform params %+v", plain.App)
	}
}

func TestGetParamsWithoutBundleOpensTheStartURL(t *testing.T) {
	for _, url := range []string{"", "https://example.com"} {
		got := GetParams(nil, url)
		if got.App != nil || got.Web == nil || got.Web.StartURL != url {
			t.Fatalf("params for %q: %+v", url, got)
		}
	}
}
