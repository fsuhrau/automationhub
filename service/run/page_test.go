package run_test

import (
	"errors"
	"math"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func key() run.Key { return run.Key{TestRef: ref(), RunID: 5} }

func detail() *models.TestRun {
	return &models.TestRun{
		Model:     models.Model{ID: 5},
		Protocols: []models.TestProtocol{{AvgFPS: math.NaN()}},
		DeviceStatus: []models.TestRunDeviceStatus{
			{DeviceID: 1, StartupTime: 100}, {DeviceID: 2, StartupTime: 100},
		},
	}
}

func TestGetBuildsTheRunPage(t *testing.T) {
	e := newEnv()
	e.repo.runDetail = detail()
	e.repo.around = run.Neighbours{Prev: 4, Next: 6}
	e.repo.startups = []models.TestRunDeviceStatus{{StartupTime: 10}, {StartupTime: 31}}
	e.runState[5] = runs.RunRunning
	e.runner.held = []uint{1}
	e.runners[5] = e.runner

	page, err := e.svc.Get(bg(), key())

	if err != nil {
		t.Fatal(err)
	}
	if page.TestRun.ID != 5 || page.PrevRunID != 4 || page.NextRunID != 6 || page.State != runs.RunRunning {
		t.Errorf("page %+v", page)
	}
	if len(page.Devices) != 1 || page.Devices[0] != 1 {
		t.Errorf("held devices %v", page.Devices)
	}
	if page.TestRun.Protocols[0].AvgFPS != 0 {
		t.Errorf("NaN average was not cleaned: %v", page.TestRun.Protocols[0].AvgFPS)
	}
	for _, s := range page.TestRun.DeviceStatus {
		if s.HistAvgStartupTime != 20 {
			t.Errorf("startup history %d, want the average 20", s.HistAvgStartupTime)
		}
	}
}

func TestGetOfARunNobodyHoldsHasNoDevicesAndNoStartupHistory(t *testing.T) {
	e := newEnv()
	e.repo.runDetail = detail()

	page, err := e.svc.Get(bg(), key())

	if err != nil || page.Devices == nil || len(page.Devices) != 0 || page.State != "" {
		t.Fatalf("page %+v, err %v", page, err)
	}
	if got := page.TestRun.DeviceStatus[0].HistAvgStartupTime; got != 0 {
		t.Errorf("startup history %d without earlier runs", got)
	}
}

func TestGetWithoutTrackerHooksNoRunIsRunning(t *testing.T) {
	e := newEnv()
	e.repo.runDetail = detail()
	svc := run.New(run.Deps{Repo: e.repo, Runners: e.runners, Summarizer: e.summary})

	page, err := svc.Get(bg(), key())

	if err != nil || page.State != runs.RunUnknown {
		t.Fatalf("page %+v, err %v", page, err)
	}
}

func TestGetErrors(t *testing.T) {
	tests := []struct {
		name string
		fail string
		want apperr.Kind
	}{
		{"test of another project", "TestInScope", apperr.KindNotFound},
		{"run of another test", "RunInTest", apperr.KindNotFound},
		{"run load", "Run", apperr.KindInternal},
		{"startup history", "StartupTimes", apperr.KindInternal},
		{"neighbours", "Neighbours", apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.runDetail = detail()
			e.repo.fail[tc.fail] = apperr.Wrap(tc.want, errors.New("x"), "x")

			page, err := e.svc.Get(bg(), key())

			assertKind(t, err, tc.want)
			if page != nil {
				t.Fatalf("page %+v with an error", page)
			}
		})
	}
}

func TestLastBuildsThePageOfTheNewestRun(t *testing.T) {
	e := newEnv()
	e.repo.lastRun = detail()
	e.repo.around = run.Neighbours{Prev: 4}

	page, err := e.svc.Last(bg(), ref())

	if err != nil || page.TestRun.ID != 5 || page.PrevRunID != 4 || page.NextRunID != 0 {
		t.Fatalf("page %+v, err %v", page, err)
	}
}

func TestLastOfATestWithoutRunsIsAnInternalError(t *testing.T) {
	e := newEnv()
	e.repo.fail["LastRun"] = apperr.NotFound("record not found")

	_, err := e.svc.Last(bg(), ref())

	assertKind(t, err, apperr.KindInternal)
	if err.Error() != "record not found" {
		t.Fatalf("message %q", err)
	}
}

func TestLastErrors(t *testing.T) {
	for _, fail := range []string{"TestInScope", "LastRun", "Neighbours"} {
		t.Run(fail, func(t *testing.T) {
			e := newEnv()
			e.repo.lastRun = detail()
			e.repo.fail[fail] = apperr.Internal("db")

			_, err := e.svc.Last(bg(), ref())

			assertKind(t, err, apperr.KindInternal)
		})
	}
}
