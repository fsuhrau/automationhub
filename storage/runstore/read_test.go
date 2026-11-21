package runstore_test

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestTestInScopeNeedsTheAppAndTheProject(t *testing.T) {
	w := newWorld(t)
	tests := []struct {
		name string
		ref  func() (testID, appID, projectID uint)
		ok   bool
	}{
		{"own test", func() (uint, uint, uint) { return w.mine.ID, w.mineApp.ID, w.mineProject.ID }, true},
		{"foreign test through the own app", func() (uint, uint, uint) { return w.foreign.ID, w.mineApp.ID, w.mineProject.ID }, false},
		{"own app of another project", func() (uint, uint, uint) { return w.mine.ID, w.mineApp.ID, w.foreProj.ID }, false},
		{"unknown test", func() (uint, uint, uint) { return 9999, w.mineApp.ID, w.mineProject.ID }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testID, appID, projectID := tc.ref()
			ref := w.ref(&models.Test{Model: models.Model{ID: testID}}, &models.App{Model: models.Model{ID: appID}}, &models.Project{Model: models.Model{ID: projectID}})

			err := w.store.TestInScope(bg(), ref)

			if tc.ok && err != nil {
				t.Fatal(err)
			}
			if !tc.ok {
				assertKind(t, err, apperr.KindNotFound, "test "+itoa(testID)+" not found")
			}
		})
	}
}

func TestRunInTestNeedsTheTest(t *testing.T) {
	w := newWorld(t)

	if err := w.store.RunInTest(bg(), w.mine.ID, w.run1.ID); err != nil {
		t.Fatal(err)
	}
	err := w.store.RunInTest(bg(), w.mine.ID, w.foreignRun.ID)
	assertKind(t, err, apperr.KindNotFound, "run "+itoa(w.foreignRun.ID)+" not found")
}

func TestTestRunsPageNewestFirstWithoutTheBundleInfo(t *testing.T) {
	w := newWorld(t)
	bundle := models.AppBinary{AppID: w.mineApp.ID, Name: "Game", Version: "1.2", Info: &app.BundleInfo{Format: "apk"}}
	w.db.Create(&bundle)
	w.db.Model(w.run2).Update("app_binary_id", bundle.ID)

	page, err := w.store.TestRuns(bg(), run.PageQuery{TestID: w.mine.ID, Limit: 1})
	if err != nil || len(page) != 1 || page[0].ID != w.run2.ID || page[0].AppBinary == nil || page[0].AppBinary.Version != "1.2" {
		t.Fatalf("page %+v, err %v", page, err)
	}
	if page[0].AppBinary.Info != nil {
		t.Error("the bundle info is large and not needed in a list")
	}
	older, err := w.store.TestRuns(bg(), run.PageQuery{TestID: w.mine.ID, Before: w.run2.ID, Limit: 5})
	if err != nil || len(older) != 1 || older[0].ID != w.run1.ID {
		t.Fatalf("older %+v, err %v", older, err)
	}
}

func TestRunEndsListsTheEndOfEveryEndedProtocol(t *testing.T) {
	w := newWorld(t)
	end := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	w.db.Model(w.proto1).Update("ended_at", end)

	got, err := w.store.RunEnds(bg(), []uint{w.run1.ID, w.run2.ID})

	if err != nil || len(got) != 1 || got[0].TestRunID != w.run1.ID || !got[0].EndedAt.Equal(end) {
		t.Fatalf("ends %+v, err %v", got, err)
	}
}

func TestLastRunAndRunLoadTheRunInDetail(t *testing.T) {
	w := newWorld(t)
	w.db.Create(&models.TestRunRecording{TestRunID: w.run2.ID, FileName: "b", StartedAt: time.Unix(200, 0)})
	w.db.Create(&models.TestRunRecording{TestRunID: w.run2.ID, FileName: "a", StartedAt: time.Unix(100, 0)})

	last, err := w.store.LastRun(bg(), w.mine.ID)
	if err != nil || last.ID != w.run2.ID || last.Test == nil || len(last.Protocols) != 1 || len(last.DeviceStatus) != 1 {
		t.Fatalf("last run %+v, err %v", last, err)
	}
	if len(last.Recordings) != 0 {
		t.Error("the last run has never listed recordings")
	}
	full, err := w.store.Run(bg(), w.run2.ID)
	if err != nil || len(full.Recordings) != 2 || full.Recordings[0].FileName != "a" {
		t.Fatalf("run %+v, err %v; want the recordings by start", full, err)
	}
}

func TestLastRunOfATestWithoutRunsAndAnUnknownRunAreNotFound(t *testing.T) {
	w := newWorld(t)
	empty := testkitTest(w)

	_, err := w.store.LastRun(bg(), empty.ID)
	assertKind(t, err, apperr.KindNotFound, "record not found")
	_, err = w.store.Run(bg(), 9999)
	assertKind(t, err, apperr.KindNotFound, "run 9999 not found")
}

func TestNeighboursAreTheRunsOfTheSameTest(t *testing.T) {
	w := newWorld(t)

	first, err := w.store.Neighbours(bg(), w.mine.ID, w.run1.ID)
	if err != nil || first != (run.Neighbours{Next: w.run2.ID}) {
		t.Fatalf("first %+v, err %v", first, err)
	}
	last, err := w.store.Neighbours(bg(), w.mine.ID, w.run2.ID)
	if err != nil || last != (run.Neighbours{Prev: w.run1.ID}) {
		t.Fatalf("last %+v, err %v", last, err)
	}
}

func TestStartupTimesIgnoreOtherAppsAndLaterRuns(t *testing.T) {
	w := newWorld(t)
	status := &models.TestRunDeviceStatus{TestRunID: w.run2.ID, DeviceID: w.device.ID}

	got, err := w.store.StartupTimes(bg(), run.StartupQuery{AppID: w.mineApp.ID, Status: status, Limit: 20})

	if err != nil || len(got) != 1 || got[0].StartupTime != 20 {
		t.Fatalf("startup times %+v, err %v; the foreign run's 1000 must not count", got, err)
	}
	limited, err := w.store.StartupTimes(bg(), run.StartupQuery{AppID: w.mineApp.ID, Status: &models.TestRunDeviceStatus{TestRunID: w.run2.ID + 100, DeviceID: w.device.ID}, Limit: 1})
	if err != nil || len(limited) != 1 || limited[0].StartupTime != 30 {
		t.Fatalf("limited %+v, err %v; want the newest earlier run", limited, err)
	}
}

func TestReadsReportADatabaseError(t *testing.T) {
	s := closed(t)
	status := &models.TestRunDeviceStatus{}
	_, e1 := s.TestRuns(bg(), run.PageQuery{TestID: 1, Limit: 1})
	_, e2 := s.RunEnds(bg(), []uint{1})
	_, e3 := s.LastRun(bg(), 1)
	_, e4 := s.Run(bg(), 1)
	_, e5 := s.Neighbours(bg(), 1, 1)
	_, e6 := s.StartupTimes(bg(), run.StartupQuery{AppID: 1, Status: status, Limit: 1})
	for i, err := range []error{e1, e2, e3, e4, e5, e6} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("read %d: %v", i, err)
		}
	}
}
