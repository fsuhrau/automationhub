package app

import (
	"context"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestBundlesListsNewestFirstWithRunCounts(t *testing.T) {
	f := newFixture(t)
	older := f.addBundle(models.AppBinary{Version: "1", Info: &hubapp.BundleInfo{Identifier: "com.app", Android: &hubapp.AndroidInfo{MinSdk: "24"}}})
	newer := f.addBundle(models.AppBinary{Version: "2"})
	f.repo.runs = []models.TestRun{{AppBinaryID: older.ID}, {AppBinaryID: older.ID}}

	got, err := f.svc.Bundles(context.Background(), f.sc)
	if err != nil || len(got) != 2 {
		t.Fatalf("bundles %+v, %v", got, err)
	}
	if got[0].ID != newer.ID || got[0].RunCount != 0 || got[1].ID != older.ID || got[1].RunCount != 2 {
		t.Errorf("order or counts wrong: %+v", got)
	}
	if got[1].Info.Identifier != "com.app" || got[1].Info.Android != nil {
		t.Errorf("list info %+v, want the summary without the details", got[1].Info)
	}
}

func TestBundlesOfAnotherProjectAreEmpty(t *testing.T) {
	f := newFixture(t)
	f.addBundle(models.AppBinary{Version: "1"})
	got, err := f.svc.Bundles(context.Background(), Scope{ProjectID: 2, AppID: 1})
	if err != nil || len(got) != 0 {
		t.Fatalf("bundles %+v, %v; want none", got, err)
	}
}

func TestBundlesErrors(t *testing.T) {
	for _, method := range []string{"Bundles", "RunCounts"} {
		f := newFixture(t)
		f.addBundle(models.AppBinary{})
		f.repo.fail[method] = errBoom
		if _, err := f.svc.Bundles(context.Background(), f.sc); err == nil {
			t.Errorf("%s: a store error must come back", method)
		}
	}
}

func TestBundleDetailWithRuns(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{Version: "1"})
	f.repo.runs = []models.TestRun{
		{Model: models.Model{ID: 1}, AppBinaryID: ref.ID, TestID: 5, Test: &models.Test{Name: "smoke"}},
		{Model: models.Model{ID: 2}, AppBinaryID: ref.ID, TestID: 6},
	}
	got, err := f.svc.Bundle(context.Background(), ref)
	if err != nil || got.RunCount != 2 || len(got.Runs) != 2 {
		t.Fatalf("detail %+v, %v", got, err)
	}
	if got.Runs[0].ID != 2 || got.Runs[0].TestName != "" || got.Runs[1].TestName != "smoke" || got.Runs[1].Status != "passed" {
		t.Errorf("runs %+v, want newest first with the test names", got.Runs)
	}
}

func TestBundleWithoutRunsHasAnEmptyList(t *testing.T) {
	f := newFixture(t)
	got, err := f.svc.Bundle(context.Background(), f.addBundle(models.AppBinary{}))
	if err != nil || got.Runs == nil || len(got.Runs) != 0 {
		t.Fatalf("runs %#v, %v; want an empty, non-nil list", got.Runs, err)
	}
}

func TestBundleErrors(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{})
	_, err := f.svc.Bundle(context.Background(), BundleRef{Scope: f.sc, ID: 99})
	wantKind(t, err, apperr.KindNotFound)
	_, err = f.svc.Bundle(context.Background(), BundleRef{Scope: Scope{ProjectID: 2, AppID: 1}, ID: ref.ID})
	wantKind(t, err, apperr.KindNotFound)
	for _, method := range []string{"RunCount", "RecentRuns", "Summarize"} {
		f.repo.fail[method] = errBoom
		if _, err := f.svc.Bundle(context.Background(), ref); err == nil {
			t.Errorf("%s: a store error must come back", method)
		}
		delete(f.repo.fail, method)
	}
}

func TestSetTags(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{Tags: "old"})
	got, err := f.svc.SetTags(context.Background(), ref, " nightly, ,rc1,nightly ")
	if err != nil || got.Tags != "nightly,rc1" || f.repo.bundles[0].Tags != "nightly,rc1" {
		t.Fatalf("tags %+v, %v; want them tidied and saved", got, err)
	}
}

func TestSetTagsErrors(t *testing.T) {
	f := newFixture(t)
	ref := f.addBundle(models.AppBinary{})
	_, err := f.svc.SetTags(context.Background(), BundleRef{Scope: f.sc, ID: 99}, "x")
	wantKind(t, err, apperr.KindNotFound)
	f.repo.fail["SaveTags"] = errBoom
	if _, err := f.svc.SetTags(context.Background(), ref, "x"); err == nil {
		t.Fatal("a store error must come back")
	}
}

func TestNormalizeTags(t *testing.T) {
	tests := map[string]string{"": "", " a , b,a ,, c": "a,b,c", "x": "x"}
	for in, want := range tests {
		if got := NormalizeTags(in); got != want {
			t.Errorf("NormalizeTags(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSummaryInfoOfNilIsNil(t *testing.T) {
	if summaryInfo(nil) != nil {
		t.Error("no info must stay no info")
	}
}
