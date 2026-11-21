package device_test

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

var _ device.CommandRunner = device.ExecRunner{}

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

func TestExecRunnerRunErrorKeepsTheOutput(t *testing.T) {
	out, err := device.ExecRunner{}.Run(context.Background(), "sh", "-c", "echo partial; echo 'Failure [INSTALL_FAILED]' >&2; exit 3")
	var cmdErr *device.CommandError
	if !errors.As(err, &cmdErr) {
		t.Fatalf("err = %v, want a *device.CommandError", err)
	}
	if strings.TrimSpace(string(out)) != "partial" || strings.TrimSpace(string(cmdErr.Stdout)) != "partial" {
		t.Errorf("stdout = %q / %q", out, cmdErr.Stdout)
	}
	if strings.TrimSpace(string(cmdErr.Stderr)) != "Failure [INSTALL_FAILED]" {
		t.Errorf("stderr = %q", cmdErr.Stderr)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Errorf("the exit error is not wrapped: %v", err)
	}
	if want := "run sh: exit status 3: Failure [INSTALL_FAILED]"; err.Error() != want {
		t.Errorf("Error() = %q, want %q", err, want)
	}
}

func TestCommandErrorWithoutStderr(t *testing.T) {
	err := &device.CommandError{Name: "adb", Stderr: []byte(" \n"), Err: errors.New("boom")}
	if err.Error() != "run adb: boom" {
		t.Errorf("Error() = %q", err)
	}
}

// syncBuffer is an io.Writer that tests read while the command writes.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func waitUntil(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestExecRunnerStartStreamingWritesStdoutAndStderr(t *testing.T) {
	var out syncBuffer
	p, err := device.ExecRunner{}.StartStreaming(context.Background(), &out, "sh", "-c", "echo to-stdout; echo to-stderr >&2")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "to-stdout") || !strings.Contains(got, "to-stderr") {
		t.Errorf("output = %q", got)
	}
}

func TestExecRunnerStartStreamingKillEndsTheProcessGroup(t *testing.T) {
	p, err := device.ExecRunner{}.StartStreaming(context.Background(), &syncBuffer{}, "sh", "-c", "trap '' INT; sleep 60")
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

func TestExecRunnerInterruptsAPlainProcess(t *testing.T) {
	p, err := device.ExecRunner{}.Start(context.Background(), "sleep", "60")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.(device.Interrupter).Interrupt(); err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err == nil {
		t.Error("an interrupted process should fail Wait")
	}
}

func TestExecRunnerStartStreamingErrors(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	for name, ctx := range map[string]context.Context{"cancelled context": cancelled, "background": context.Background()} {
		t.Run(name, func(t *testing.T) {
			p, err := device.ExecRunner{}.StartStreaming(ctx, &syncBuffer{}, "/no/such/command-g33")
			if err == nil || !strings.Contains(err.Error(), "start /no/such/command-g33") {
				t.Errorf("err = %v", err)
			}
			if p != nil {
				t.Errorf("a failed start must return a nil Process, got %v", p)
			}
		})
	}
}
