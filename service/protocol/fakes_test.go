package protocol_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func bg() context.Context { return context.Background() }

// fakeRepo answers every call with its fields and records the queries.
type fakeRepo struct {
	fail       map[string]error
	run        *models.TestRun
	history    []models.TestProtocol
	protocol   models.TestProtocol
	candidates []models.TestProtocol
	baseline   models.TestProtocol
	tags       []string

	historyOf      *models.TestProtocol
	candidateQuery protocol.CandidateQuery
	baselineQuery  protocol.BaselineQuery
}

func (f *fakeRepo) RunWithProtocol(context.Context, uint, uint) (*models.TestRun, error) {
	return f.run, f.fail["RunWithProtocol"]
}

func (f *fakeRepo) History(_ context.Context, _ uint, current *models.TestProtocol) ([]models.TestProtocol, error) {
	f.historyOf = current
	return f.history, f.fail["History"]
}

func (f *fakeRepo) Protocol(context.Context, protocol.Key) (*models.TestProtocol, error) {
	return &f.protocol, f.fail["Protocol"]
}

func (f *fakeRepo) Candidates(_ context.Context, q protocol.CandidateQuery) ([]models.TestProtocol, error) {
	f.candidateQuery = q
	return f.candidates, f.fail["Candidates"]
}

func (f *fakeRepo) Baseline(_ context.Context, q protocol.BaselineQuery) (*models.TestProtocol, error) {
	f.baselineQuery = q
	return &f.baseline, f.fail["Baseline"]
}

func (f *fakeRepo) BundleTags(context.Context, uint) ([]string, error) {
	return f.tags, f.fail["BundleTags"]
}

// scopeState is the answer every run check gets.
type scopeState struct{ err error }

func (f *scopeState) CheckRun(context.Context, testsvc.TestRef, uint) error { return f.err }

type env struct {
	svc   *protocol.Service
	repo  *fakeRepo
	scope *scopeState
}

func newEnv() *env {
	e := &env{repo: &fakeRepo{fail: map[string]error{}}, scope: &scopeState{}}
	e.svc = protocol.New(protocol.Deps{
		Repo: e.repo, Scope: e.scope,
		DeviceName: func(d *models.Device) string { return "device " + d.Name },
	})
	return e
}

func key() protocol.Key {
	return protocol.Key{TestRef: testsvc.TestRef{Scope: testsvc.Scope{ProjectID: 1, AppID: 2}, TestID: 3}, RunID: 4, ProtocolID: 5}
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("error kind %v (%v), want %v", got, err, want)
	}
}
