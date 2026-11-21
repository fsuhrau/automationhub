package runstore_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// failQueries makes every query whose SQL contains the fragment fail.
func failQueries(db *gorm.DB, fragment string) {
	fail := func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), fragment) {
			_ = tx.AddError(errors.New("query failed"))
		}
	}
	_ = db.Callback().Query().After("gorm:query").Register("fail:"+fragment, fail)
	_ = db.Callback().Row().After("gorm:row").Register("fail:"+fragment, fail)
}

func TestReportRunLoadsTheRunWithItsTestBundleAndProtocolsInOrder(t *testing.T) {
	w := newWorld(t)
	bundle := &models.AppBinary{AppID: w.mineApp.ID, Name: "Game"}
	w.db.Create(bundle)
	w.db.Model(w.run1).Update("app_binary_id", bundle.ID)
	w.db.Create(&models.ProtocolKey{TestProtocolID: w.proto1.ID, Key: "player", Value: "1"})
	second := w.protocolWithout(w.run1)

	got, err := w.store.ReportRun(bg(), w.run1.ID)

	if err != nil || got.Test == nil || got.AppBinary == nil || len(got.Protocols) != 2 {
		t.Fatalf("report run %+v, err %v", got, err)
	}
	if got.Protocols[0].ID != w.proto1.ID || got.Protocols[1].ID != second.ID {
		t.Errorf("protocols %d, %d, want by id", got.Protocols[0].ID, got.Protocols[1].ID)
	}
	if got.Protocols[0].Device == nil || len(got.Protocols[0].Keys) != 1 {
		t.Errorf("protocol %+v has no device or keys", got.Protocols[0])
	}
}

func TestReportRunOfAMissingRunIsNotFound(t *testing.T) {
	w := newWorld(t)

	_, err := w.store.ReportRun(bg(), 9999)

	assertKind(t, err, apperr.KindNotFound, "not found: test run 9999")
}

func TestLogTailListsTheNewestEntriesFirst(t *testing.T) {
	w := newWorld(t)
	for _, line := range []string{"a", "b", "c"} {
		w.db.Create(&models.TestRunLogEntry{TestRunID: w.run1.ID, Level: "info", Log: line})
	}
	w.db.Create(&models.TestRunLogEntry{TestRunID: w.run2.ID, Level: "info", Log: "other run"})

	got, err := w.store.LogTail(bg(), w.run1.ID, 2)

	if err != nil || len(got) != 2 || got[0].Log != "c" || got[1].Log != "b" {
		t.Fatalf("tail %+v, err %v; want c, b", got, err)
	}
}

func TestProtocolProgressCountsTheProtocolsAndTheEndedOnes(t *testing.T) {
	w := newWorld(t)
	ended := testkit.Protocol(w.db, func(p *models.TestProtocol) { p.TestRunID = w.run1.ID })
	w.db.Model(ended).Update("ended_at", ended.StartedAt)

	got, err := w.store.ProtocolProgress(bg(), w.run1.ID)

	if err != nil || got.Total != 2 || got.Ended != 1 {
		t.Fatalf("progress %+v, err %v; want 1 of 2 ended", got, err)
	}
}

func TestProtocolProgressReportsAFailingCountOfTheEndedProtocols(t *testing.T) {
	w := newWorld(t)
	failQueries(w.db, "ended_at IS NOT NULL")

	_, err := w.store.ProtocolProgress(bg(), w.run1.ID)

	assertKind(t, err, apperr.KindInternal, "")
}

func TestReportReadsReportADatabaseError(t *testing.T) {
	s := closed(t)
	_, e1 := s.ReportRun(bg(), 1)
	_, e2 := s.LogTail(bg(), 1, 1)
	_, e3 := s.ProtocolProgress(bg(), 1)
	for i, err := range []error{e1, e2, e3} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("read %d: %v", i, err)
		}
	}
}
