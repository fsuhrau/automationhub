package runstore_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestProtocolDigestSumsUpTheProtocol(t *testing.T) {
	w := newWorld(t)
	entries(w, "start", "boom", "fine", "crash")
	w.db.Create(&models.ProtocolEntry{TestProtocolID: w.proto1.ID, Level: "info", Message: "shot", Data: "a.png"})
	w.db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: w.proto1.ID, Checkpoint: "end", FPS: 30, MEM: 8, CPU: 4})
	child := w.protocolWithout(w.run1)
	w.db.Model(child).Update("parent_test_protocol_id", w.proto1.ID)

	got, err := w.store.ProtocolDigest(bg(), w.proto1.ID, 50)

	if err != nil {
		t.Fatal(err)
	}
	if got.Entries != 5 || len(got.Errors) != 2 || got.Errors[0].Message != "boom" || len(got.Data) != 1 {
		t.Errorf("entries %d, errors %+v, data %+v", got.Entries, got.Errors, got.Data)
	}
	if p := got.Performance; p.Samples != 2 || p.AvgFPS != 20 || p.MinFPS != 10 || p.MaxMEM != 8 || p.AvgCPU != 2 {
		t.Errorf("performance %+v", p)
	}
	if len(got.Children) != 1 || got.Children[0].ID != child.ID {
		t.Errorf("children %+v, want the child protocol", got.Children)
	}
}

func TestProtocolDigestListsAtMostLimitErrors(t *testing.T) {
	w := newWorld(t)
	entries(w, "a", "b", "c", "d")

	got, err := w.store.ProtocolDigest(bg(), w.proto1.ID, 1)

	if err != nil || len(got.Errors) != 1 || got.Errors[0].Message != "b" {
		t.Fatalf("errors %+v, err %v; want the first one", got.Errors, err)
	}
}

func TestDataEntriesAreThoseWithDataByID(t *testing.T) {
	w := newWorld(t)
	entries(w, "no data")
	w.db.Create(&models.ProtocolEntry{TestProtocolID: w.proto1.ID, Message: "one", Data: "a.png"})
	w.db.Create(&models.ProtocolEntry{TestProtocolID: w.proto1.ID, Message: "two", Data: "b.json"})
	w.db.Create(&models.ProtocolEntry{TestProtocolID: w.proto2.ID, Message: "other", Data: "c.png"})

	got, err := w.store.DataEntries(bg(), w.proto1.ID)

	if err != nil || len(got) != 2 || got[0].Message != "one" || got[1].Message != "two" {
		t.Fatalf("data entries %+v, err %v", got, err)
	}
}

func TestProtocolDigestReportsEachFailingQuery(t *testing.T) {
	tests := map[string]string{
		"entry count":  "count(*)",
		"errors":       "level IN",
		"performance":  "protocol_performance_entries",
		"children":     "parent_test_protocol_id",
		"data entries": "data <>",
	}
	for name, fragment := range tests {
		t.Run(name, func(t *testing.T) {
			w := newWorld(t)
			failQueries(w.db, fragment)

			_, err := w.store.ProtocolDigest(bg(), w.proto1.ID, 5)

			assertKind(t, err, apperr.KindInternal, "")
		})
	}
}

func TestDigestReadsReportADatabaseError(t *testing.T) {
	_, err := closed(t).DataEntries(bg(), 1)

	assertKind(t, err, apperr.KindInternal, "")
}
