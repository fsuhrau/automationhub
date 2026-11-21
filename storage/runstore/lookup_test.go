package runstore_test

import (
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestProjectProtocolLoadsItWithDeviceAndKeys(t *testing.T) {
	w := newWorld(t)
	w.db.Create(&models.ProtocolKey{TestProtocolID: w.proto1.ID, Key: "player", Value: "1"})

	got, err := w.store.ProjectProtocol(bg(), w.mineProject.ID, w.proto1.ID)

	if err != nil || got.Device == nil || len(got.Keys) != 1 {
		t.Fatalf("protocol %+v, err %v", got, err)
	}
	_, err = w.store.ProjectProtocol(bg(), w.mineProject.ID, w.foreP.ID)
	assertKind(t, err, apperr.KindNotFound, fmt.Sprintf("not found: protocol %d", w.foreP.ID))
}

func entries(w *world, messages ...string) {
	for i, m := range messages {
		level := "info"
		if i%2 == 1 {
			level = "error"
		}
		w.db.Create(&models.ProtocolEntry{TestProtocolID: w.proto1.ID, Level: level, Message: m})
	}
}

func TestProtocolLogsFilterAndPage(t *testing.T) {
	w := newWorld(t)
	entries(w, "start", "50% done", "a_b", "FAILED hard", "end")
	tests := []struct {
		name  string
		q     run.LogQuery
		total int64
		want  []string
	}{
		{"all", run.LogQuery{Limit: 10}, 5, []string{"start", "50% done", "a_b", "FAILED hard", "end"}},
		{"paged", run.LogQuery{Offset: 1, Limit: 2}, 5, []string{"50% done", "a_b"}},
		{"by level, any case", run.LogQuery{Level: "ERROR", Limit: 10}, 2, []string{"50% done", "FAILED hard"}},
		{"containing a percent sign", run.LogQuery{Contains: "50%", Limit: 10}, 1, []string{"50% done"}},
		{"containing an underscore", run.LogQuery{Contains: "a_b", Limit: 10}, 1, []string{"a_b"}},
		{"containing nothing that matches literally", run.LogQuery{Contains: "a%b", Limit: 10}, 0, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.q.ProtocolID = w.proto1.ID

			got, total, err := w.store.ProtocolLogs(bg(), tc.q)

			if err != nil || total != tc.total || len(got) != len(tc.want) {
				t.Fatalf("got %d entries of %d, err %v; want %v of %d", len(got), total, err, tc.want, tc.total)
			}
			for i, e := range got {
				if e.Message != tc.want[i] {
					t.Errorf("entry %d is %q, want %q", i, e.Message, tc.want[i])
				}
			}
		})
	}
}

func TestProtocolLogsReportADatabaseError(t *testing.T) {
	_, _, err := closed(t).ProtocolLogs(bg(), run.LogQuery{Limit: 1})

	assertKind(t, err, apperr.KindInternal, "")
}
