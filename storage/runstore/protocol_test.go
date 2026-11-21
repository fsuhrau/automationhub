package runstore_test

import (
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestRunWithProtocolLoadsOnlyThatProtocol(t *testing.T) {
	w := newWorld(t)
	second := w.protocol(w.run1, 5)
	w.db.Create(&models.ProtocolEntry{TestProtocolID: second.ID, Message: "hello"})

	got, err := w.store.RunWithProtocol(bg(), w.run1.ID, second.ID)

	if err != nil || len(got.Protocols) != 1 || got.Protocols[0].ID != second.ID || len(got.Protocols[0].Entries) != 1 ||
		got.Protocols[0].Device == nil || len(got.Protocols[0].Performance) != 1 || got.Test == nil {
		t.Fatalf("run %+v, err %v", got, err)
	}
	none, err := w.store.RunWithProtocol(bg(), w.run1.ID, 9999)
	if err != nil || len(none.Protocols) != 0 {
		t.Fatalf("a protocol that is not in the run: %+v, err %v", none, err)
	}
	_, err = w.store.RunWithProtocol(bg(), 9999, 1)
	assertKind(t, err, apperr.KindNotFound, "run 9999 not found")
}

func TestHistoryHasEarlierMeasuredProtocolsOfTheAppOnTheSameDevice(t *testing.T) {
	w := newWorld(t)
	other := testkitDevice(w)
	w.protocol(w.run2, 50)              // the run itself, a second protocol
	w.protocolWithout(w.run1)           // no performance entries
	w.protocolOn(w.run1, other.ID, 999) // another device
	w.protocol(w.run(w.mine, 1), 70)    // a later run
	current := w.proto2

	got, err := w.store.History(bg(), w.mineApp.ID, current)

	if err != nil || len(got) != 1 || got[0].ID != w.proto1.ID || len(got[0].Performance) != 1 {
		t.Fatalf("history %+v, err %v; want only the earlier measured protocol of the own app", got, err)
	}
}

func TestHistoryIsLimited(t *testing.T) {
	w := newWorld(t)
	for i := 0; i < 25; i++ {
		r := w.run(w.mine, 1)
		w.protocol(r, 1)
	}
	last := w.protocol(w.run(w.mine, 1), 1)

	got, err := w.store.History(bg(), w.mineApp.ID, last)

	if err != nil || len(got) != 20 {
		t.Fatalf("history of %d protocols, err %v; want the newest 20", len(got), err)
	}
}

func TestProtocolNeedsItsRunTestAppAndProject(t *testing.T) {
	w := newWorld(t)
	got, err := w.store.Protocol(bg(), w.key(w.run1, w.proto1))
	if err != nil || got.ID != w.proto1.ID {
		t.Fatalf("protocol %+v, err %v", got, err)
	}
	otherProject := w.key(w.run1, w.proto1)
	otherProject.ProjectID = w.foreProj.ID
	keys := map[string]protocol.Key{
		"foreign protocol": w.key(w.foreignRun, w.foreP),
		"wrong run":        w.key(w.run2, w.proto1),
		"other project":    otherProject,
	}
	for name, key := range keys {
		t.Run(name, func(t *testing.T) {
			_, err := w.store.Protocol(bg(), key)

			assertKind(t, err, apperr.KindNotFound, fmt.Sprintf("protocol %d not found in run %d", key.ProtocolID, key.RunID))
		})
	}
}

func TestProtocolReadsReportADatabaseError(t *testing.T) {
	s := closed(t)
	_, e1 := s.RunWithProtocol(bg(), 1, 1)
	_, e2 := s.History(bg(), 1, &models.TestProtocol{})
	_, e3 := s.Protocol(bg(), protocol.Key{})
	for i, err := range []error{e1, e2, e3} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("read %d: %v", i, err)
		}
	}
}
