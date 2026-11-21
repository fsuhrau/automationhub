package run_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestProtocolDetailHasTheProtocolAndItsDigest(t *testing.T) {
	e := newEnv()
	e.repo.protocol = &models.TestProtocol{Model: models.Model{ID: 5}, TestName: "S/A"}
	e.repo.digest = run.Digest{Entries: 9, Performance: run.Performance{Samples: 2}}

	got, err := e.svc.ProtocolDetail(bg(), 1, 5)

	if err != nil || got.Protocol.TestName != "S/A" || got.Entries != 9 || got.Performance.Samples != 2 {
		t.Fatalf("detail %+v, err %v", got, err)
	}
	if e.repo.digestFor != 5 || e.repo.digestMax != run.DigestErrors {
		t.Errorf("digest of protocol %d with %d errors, want 5 and %d", e.repo.digestFor, e.repo.digestMax, run.DigestErrors)
	}
}

func TestProtocolDetailFailsWhereTheRepoFails(t *testing.T) {
	for _, method := range []string{"ProjectProtocol", "ProtocolDigest"} {
		t.Run(method, func(t *testing.T) {
			e := newEnv()
			e.repo.protocol = &models.TestProtocol{}
			e.repo.fail[method] = apperr.NotFound(method)

			_, err := e.svc.ProtocolDetail(bg(), 1, 5)

			assertKind(t, err, apperr.KindNotFound)
		})
	}
}

func TestProtocolDataListsTheEntriesWithData(t *testing.T) {
	e := newEnv()
	e.repo.protocol = &models.TestProtocol{}
	e.repo.dataList = []models.ProtocolEntry{{Data: "a.png"}}

	got, err := e.svc.ProtocolData(bg(), 1, 5)

	if err != nil || len(got) != 1 || got[0].Data != "a.png" {
		t.Fatalf("data %+v, err %v", got, err)
	}
}

func TestProtocolDataFailsWhereTheRepoFails(t *testing.T) {
	for _, method := range []string{"ProjectProtocol", "DataEntries"} {
		t.Run(method, func(t *testing.T) {
			e := newEnv()
			e.repo.protocol = &models.TestProtocol{}
			e.repo.fail[method] = apperr.NotFound(method)

			_, err := e.svc.ProtocolData(bg(), 1, 5)

			assertKind(t, err, apperr.KindNotFound)
		})
	}
}
