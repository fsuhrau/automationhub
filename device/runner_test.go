package device_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

var _ device.CommandRunner = device.ExecRunner{}

// The Run signature of testkit.FakeRunner is the one of device.CommandRunner.
// Its Start returns testkit.Process, which has the methods of device.Process
// but is another type, so the fake needs an adapter or testkit.Process must
// become an alias of device.Process.
var _ func(context.Context, string, ...string) ([]byte, error) = (*testkit.FakeRunner)(nil).Run
var _ func(context.Context, string, ...string) (testkit.Process, error) = (*testkit.FakeRunner)(nil).Start
var _ device.Process = testkit.Process(nil)

func TestExecRunnerRun(t *testing.T) {
	out, err := device.ExecRunner{}.Run(context.Background(), "echo", "hello")
	if err != nil || strings.TrimSpace(string(out)) != "hello" {
		t.Errorf("Run(echo hello) = %q, %v", out, err)
	}
}

func TestExecRunnerRunErrors(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name string
		ctx  context.Context
		cmd  string
	}{
		{"exit status", context.Background(), "false"},
		{"unknown command", context.Background(), "no-such-command-g12"},
		{"cancelled context", cancelled, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := (device.ExecRunner{}).Run(tt.ctx, tt.cmd); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestExecRunnerStart(t *testing.T) {
	p, err := device.ExecRunner{}.Start(context.Background(), "true")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err != nil {
		t.Errorf("Wait() = %v", err)
	}
}

func TestExecRunnerStartKill(t *testing.T) {
	p, err := device.ExecRunner{}.Start(context.Background(), "sleep", "30")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err == nil {
		t.Error("a killed process should fail Wait")
	}
}

func TestExecRunnerStartKillsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p, err := device.ExecRunner{}.Start(ctx, "sleep", "30")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	done := make(chan error, 1)
	go func() { done <- p.Wait() }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("a cancelled process should fail Wait")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("process was not killed")
	}
}

func TestExecRunnerStartUnknownCommand(t *testing.T) {
	if _, err := (device.ExecRunner{}).Start(context.Background(), "no-such-command-g12"); err == nil {
		t.Error("expected an error")
	}
}
