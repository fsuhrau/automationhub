package run_test

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
)

func TestCancelStopsTheRunnerOfARunOfTheProject(t *testing.T) {
	e := newEnv()
	e.runners[7] = e.runner

	err := e.svc.Cancel(bg(), 1, 7)

	if err != nil || e.runner.cancelID != "7" {
		t.Fatalf("err %v, cancelled %q", err, e.runner.cancelID)
	}
	if _, active := e.runners[7]; active {
		t.Fatal("the cancelled run is still registered")
	}
}

func TestCancelOfARunOfAnotherProjectIsNotFound(t *testing.T) {
	e := newEnv()
	e.runners[7] = e.runner
	e.repo.fail["RunInProject"] = apperr.NotFound("not found: test run 7")

	err := e.svc.Cancel(bg(), 2, 7)

	assertKind(t, err, apperr.KindNotFound)
	if e.runner.cancelID != "" {
		t.Fatal("a run of another project was cancelled")
	}
}

func TestCancelOfARunThatIsNotRunningIsNotFound(t *testing.T) {
	e := newEnv()

	err := e.svc.Cancel(bg(), 1, 7)

	assertKind(t, err, apperr.KindNotFound)
	if err.Error() != "not found: no active runner for test run 7" {
		t.Fatalf("message %q", err)
	}
}

func TestCancelKeepsTheRunnerWhenCancelFails(t *testing.T) {
	e := newEnv()
	e.runners[7] = e.runner
	e.runner.failOn, e.runner.err = "Cancel", errors.New("stuck")

	err := e.svc.Cancel(bg(), 1, 7)

	assertKind(t, err, apperr.KindInternal)
	if _, active := e.runners[7]; !active {
		t.Fatal("a runner that could not be cancelled was forgotten")
	}
}
