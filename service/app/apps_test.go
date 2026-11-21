package app

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != kind {
		t.Fatalf("error %v has kind %d, want %d", err, got, kind)
	}
}

func TestAppsListsTheProjectsApps(t *testing.T) {
	f := newFixture(t)
	got, err := f.svc.Apps(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].Name != "App" {
		t.Fatalf("apps %+v, %v; want only the app of project 1", got, err)
	}
	f.repo.fail["Apps"] = errBoom
	if _, err := f.svc.Apps(context.Background(), 1); err == nil {
		t.Fatal("a store error must come back")
	}
}

func TestAppReturnsAnAppOfTheProjectOnly(t *testing.T) {
	f := newFixture(t)

	got, err := f.svc.App(context.Background(), f.sc)
	if err != nil || got.Name != "App" {
		t.Fatalf("app %+v, %v; want the app of the scope", got, err)
	}
	_, err = f.svc.App(context.Background(), Scope{ProjectID: 1, AppID: 2})
	wantKind(t, err, apperr.KindNotFound)
}

func TestCreateApp(t *testing.T) {
	tests := []struct {
		name string
		in   AppInput
		kind apperr.Kind
	}{
		{"android app", AppInput{Name: " Game ", Identifier: " com.game ", Platform: models.PlatformTypeAndroid}, apperr.KindNone},
		{"windows needs no identifier", AppInput{Name: "Game", Platform: models.PlatformTypeWindows}, apperr.KindNone},
		{"no name", AppInput{Name: " ", Platform: models.PlatformTypeWindows}, apperr.KindInvalid},
		{"no identifier", AppInput{Name: "Game", Identifier: " ", Platform: models.PlatformTypeiOS}, apperr.KindInvalid},
		{"unknown platform", AppInput{Name: "Game", Platform: 42}, apperr.KindInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			got, err := f.svc.CreateApp(context.Background(), 1, tt.in)
			wantKind(t, err, tt.kind)
			if err == nil && (got.ProjectID != 1 || got.Name != "Game" || got.ID == 0) {
				t.Errorf("created %+v, want a trimmed app of project 1", got)
			}
		})
	}
}

func TestCreateAppStoreError(t *testing.T) {
	f := newFixture(t)
	f.repo.fail["CreateApp"] = errBoom
	if _, err := f.svc.CreateApp(context.Background(), 1, AppInput{Name: "x", Platform: models.PlatformTypeWeb}); err == nil {
		t.Fatal("a store error must come back")
	}
}

func TestUpdateApp(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{"the identifier stays", "com.app", "com.app"},
		{"the default identifier is replaced once", defaultIdentifier, "com.new"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.repo.apps[0].Identifier = tt.identifier
			got, err := f.svc.UpdateApp(context.Background(), f.sc, AppInput{Name: "Renamed", Identifier: "com.new"})
			if err != nil || got.Name != "Renamed" || got.Identifier != tt.want {
				t.Fatalf("updated %+v, %v; want name Renamed, identifier %s", got, err, tt.want)
			}
			if f.repo.apps[0].Name != "Renamed" {
				t.Error("the change was not saved")
			}
		})
	}
}

func TestUpdateAppErrors(t *testing.T) {
	f := newFixture(t)
	_, err := f.svc.UpdateApp(context.Background(), Scope{ProjectID: 2, AppID: 1}, AppInput{Name: "x"})
	wantKind(t, err, apperr.KindNotFound)
	f.repo.fail["SaveApp"] = errBoom
	if _, err := f.svc.UpdateApp(context.Background(), f.sc, AppInput{Name: "x"}); err == nil {
		t.Fatal("a store error must come back")
	}
}
