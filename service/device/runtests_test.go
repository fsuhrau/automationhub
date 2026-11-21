package device_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"
	svc "github.com/fsuhrau/automationhub/service/device"
)

func TestRunTests(t *testing.T) {
	tests := []struct {
		name      string
		testName  string
		id        string
		connected bool
		app       bool
		lockErr   error
		execErr   error
		want      apperr.Kind
		executed  bool
	}{
		{"runs", "Suite Method", "1", true, true, nil, nil, apperr.KindNone, true},
		{"invalid signature", "x", "1", true, true, nil, nil, apperr.KindInvalid, false},
		{"empty signature", "", "1", true, true, nil, nil, apperr.KindInvalid, false},
		{"unknown device", "A b", "2", true, true, nil, nil, apperr.KindNotFound, false},
		{"not connected", "A b", "1", false, true, nil, nil, apperr.KindNotFound, false},
		{"lock fails", "A b", "1", true, true, errBoom, nil, apperr.KindNotFound, false},
		{"no app connected", "A b", "1", true, false, nil, nil, apperr.KindNotFound, false},
		{"execution fails", "A b", "1", true, true, nil, errBoom, apperr.KindInternal, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(stored(1, "dev-1"))
			dev := testkit.NewFakeDevice("dev-1")
			dev.AppConnected = tc.app
			dev.FailWith("Lock", tc.lockErr)
			e.executor.err = tc.execErr
			if tc.connected {
				e.pool.Add("android_device", dev)
			}

			err := e.svc.RunTests(context.Background(), e.project, svc.TestRequest{DeviceID: tc.id, TestName: tc.testName, Env: "k=v"})

			assertKind(t, err, tc.want)
			if got := len(e.executor.tests) == 1; got != tc.executed {
				t.Fatalf("executed = %v, want %v", got, tc.executed)
			}
			if dev.IsLocked() {
				t.Fatal("device left locked")
			}
		})
	}
}

func TestRunTestsStartsTheTestAfterAReset(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	dev := testkit.NewFakeDevice("dev-1")
	dev.AppConnected = true
	e.pool.Add("android_device", dev)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := e.svc.RunTests(ctx, e.project, svc.TestRequest{DeviceID: "1", TestName: "Suite Method", Env: "k=v"}); err != nil {
		t.Fatal(err)
	}

	want := action.TestStart{Class: "Suite", Method: "Method", Env: map[string]string{"raw": "k=v"}}
	if len(e.executor.tests) != 1 || e.executor.tests[0].Class != want.Class || e.executor.tests[0].Method != want.Method ||
		e.executor.tests[0].Env["raw"] != "k=v" {
		t.Fatalf("started %+v", e.executor.tests)
	}
	if len(e.pool.CallsTo("SendAction")) != 1 {
		t.Fatalf("expected the reset action, got %v", e.pool.Calls())
	}
	if e.executor.ctx.Err() != nil {
		t.Fatal("a client that went away stopped the run")
	}
}
