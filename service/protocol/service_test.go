package protocol_test

import (
	"math"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func runWith(protocols ...models.TestProtocol) *models.TestRun {
	return &models.TestRun{Model: models.Model{ID: 4}, Protocols: protocols}
}

func TestPageComparesTheProtocolWithItsHistory(t *testing.T) {
	e := newEnv()
	e.repo.run = runWith(models.TestProtocol{TestName: "S/A", AvgFPS: math.NaN()})
	e.repo.history = []models.TestProtocol{{AvgFPS: 10}, {AvgFPS: 30}}

	run, err := e.svc.Page(bg(), key())

	if err != nil {
		t.Fatal(err)
	}
	got := run.Protocols[0]
	if got.AvgFPS != 0 || got.HistAvgFPS != 20 || len(got.TestProtocolHistory) != 2 {
		t.Fatalf("protocol %+v", got)
	}
	if e.repo.historyOf != &run.Protocols[0] {
		t.Fatal("the history was not asked for the loaded protocol")
	}
}

func TestPageOfAProtocolThatIsNotInTheRunHasNoHistory(t *testing.T) {
	e := newEnv()
	e.repo.run = runWith()

	run, err := e.svc.Page(bg(), key())

	if err != nil || len(run.Protocols) != 0 || e.repo.historyOf != nil {
		t.Fatalf("run %+v, err %v, history asked %v", run, err, e.repo.historyOf)
	}
}

func TestPageErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*env)
		want  apperr.Kind
	}{
		{"run of another test", func(e *env) { e.scope.err = apperr.NotFound("run 4 not found") }, apperr.KindNotFound},
		{"run load", func(e *env) { e.repo.fail["RunWithProtocol"] = apperr.Internal("db") }, apperr.KindInternal},
		{"history load", func(e *env) {
			e.repo.run = runWith(models.TestProtocol{})
			e.repo.fail["History"] = apperr.Internal("db")
		}, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			tc.setup(e)

			run, err := e.svc.Page(bg(), key())

			assertKind(t, err, tc.want)
			if run != nil {
				t.Fatalf("run %+v with an error", run)
			}
		})
	}
}
