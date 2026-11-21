package test_test

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/runs"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func mine() service.Scope    { return service.Scope{ProjectID: 1, AppID: 2} }
func foreign() service.Scope { return service.Scope{ProjectID: 9, AppID: 2} }

// constErr is an error that is comparable and constant.
type constErr string

func (e constErr) Error() string { return string(e) }

const boom = constErr("boom")

func unityTest(id uint) models.Test {
	return models.Test{
		Model: models.Model{ID: id}, AppID: 2, Name: "Smoke",
		TestConfig: models.TestConfig{
			Model: models.Model{ID: 5}, Type: models.TestTypeUnity,
			Unity: &models.TestConfigUnity{Model: models.Model{ID: 6}},
		},
	}
}

func newFake() *fakeRepo {
	return &fakeRepo{scope: mine(), tests: []models.Test{unityTest(1)}, visible: 1, failWith: map[string]error{}}
}

func ref(sc service.Scope, id uint) service.TestRef { return service.TestRef{Scope: sc, TestID: id} }

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != kind {
		t.Fatalf("error %v has kind %d, want %d", err, got, kind)
	}
}

func TestGetEnforcesTheProjectScope(t *testing.T) {
	tests := []struct {
		name string
		sc   service.Scope
		id   uint
		kind apperr.Kind
	}{
		{"own test", mine(), 1, apperr.KindNone},
		{"unknown test", mine(), 7, apperr.KindNotFound},
		{"another project's test", foreign(), 1, apperr.KindNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.New(service.Deps{Repo: newFake()})

			test, err := svc.Get(bg(), ref(tc.sc, tc.id))

			wantKind(t, err, tc.kind)
			if tc.kind == apperr.KindNone && test.ID != 1 {
				t.Fatalf("test %+v", test)
			}
		})
	}
}

func TestListAddsRecentRunSummaries(t *testing.T) {
	repo := newFake()
	repo.runs = []models.TestRun{{Model: models.Model{ID: 10}}}
	repo.results = []service.ResultCount{{TestRunID: 10, TestResult: models.TestResultSuccess, N: 2}}
	svc := service.New(service.Deps{Repo: repo})

	list, err := svc.List(bg(), mine())

	if err != nil || len(list) != 1 || len(list[0].RecentRuns) != 1 {
		t.Fatalf("list %+v, err %v", list, err)
	}
	if got := list[0].RecentRuns[0]; got.ID != 10 || got.Passed != 2 || got.Status != "passed" {
		t.Fatalf("summary %+v", got)
	}
}

func TestListOfAnotherProjectIsEmpty(t *testing.T) {
	svc := service.New(service.Deps{Repo: newFake()})

	list, err := svc.List(bg(), foreign())

	if err != nil || len(list) != 0 {
		t.Fatalf("list %+v, err %v", list, err)
	}
}

func TestListReportsRepoFailures(t *testing.T) {
	for _, step := range []string{"Tests", "RecentRuns", "RunResults", "RunLogErrors"} {
		t.Run(step, func(t *testing.T) {
			repo := newFake()
			repo.runs = []models.TestRun{{Model: models.Model{ID: 10}}}
			repo.failWith[step] = boom
			svc := service.New(service.Deps{Repo: repo})

			_, err := svc.List(bg(), mine())

			if !errors.Is(err, boom) {
				t.Fatalf("err %v, want it to wrap the repo error", err)
			}
		})
	}
}

func TestSummariesStatusAndCounts(t *testing.T) {
	tests := []struct {
		name      string
		state     runs.RunState
		results   []service.ResultCount
		logErrors []service.LogErrorCount
		want      string
	}{
		{"running", runs.RunRunning, []service.ResultCount{{TestRunID: 1, TestResult: models.TestResultFailed, N: 1}}, nil, "running"},
		{"failed", runs.RunDone, []service.ResultCount{{TestRunID: 1, TestResult: models.TestResultSuccess, N: 3}, {TestRunID: 1, TestResult: models.TestResultFailed, N: 1}, {TestRunID: 1, TestResult: models.TestResultOpen, N: 1}}, nil, "failed"},
		{"unstable", runs.RunUnknown, []service.ResultCount{{TestRunID: 1, TestResult: models.TestResultUnstable, N: 1}}, nil, "unstable"},
		{"interrupted", runs.RunUnknown, []service.ResultCount{{TestRunID: 1, TestResult: models.TestResultOpen, N: 2}}, nil, "interrupted"},
		{"passed", runs.RunDone, []service.ResultCount{{TestRunID: 1, TestResult: models.TestResultSuccess, N: 3}}, []service.LogErrorCount{{TestRunID: 1, N: 2}}, "passed"},
		{"not run", runs.RunUnknown, nil, []service.LogErrorCount{{TestRunID: 1, N: 2}}, "not_run"},
		{"empty", runs.RunDone, nil, nil, "empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFake()
			repo.results, repo.logErrors = tc.results, tc.logErrors
			svc := service.New(service.Deps{Repo: repo, RunState: func(uint) runs.RunState { return tc.state }})

			got, err := svc.Summaries(bg(), []models.TestRun{{Model: models.Model{ID: 1}}})

			if err != nil || got[0].Status != tc.want || got[0].State != tc.state {
				t.Fatalf("summary %+v, err %v; want status %s", got, err, tc.want)
			}
		})
	}
}

func TestSummariesWithoutRunsSkipsTheRepo(t *testing.T) {
	repo := newFake()

	got, err := service.New(service.Deps{Repo: repo}).Summaries(bg(), nil)

	if err != nil || len(got) != 0 || len(repo.calls) != 0 {
		t.Fatalf("summaries %v, err %v, calls %v", got, err, repo.calls)
	}
}

func TestSummariesWithoutRunStateAreNotRunning(t *testing.T) {
	svc := service.New(service.Deps{Repo: newFake()})

	got, _ := svc.Summaries(bg(), []models.TestRun{{Model: models.Model{ID: 1}}})

	if got[0].State != runs.RunUnknown {
		t.Fatalf("state %q", got[0].State)
	}
}
