package device

import (
	"bytes"
	"context"
	"io"
)

// Process is a started command.
type Process interface {
	Wait() error
	Kill() error
}

// CommandRunner runs external commands for the platform packages.
type CommandRunner interface {
	// Run runs the command and returns its stdout. When the command fails
	// the error is a *CommandError with the stdout and stderr.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	// Start starts the command in the background; cancelling ctx kills it.
	Start(ctx context.Context, name string, args ...string) (Process, error)
	// StartStreaming starts the command in the background and writes its
	// stdout and stderr, interleaved, to out. It is for commands whose output
	// is needed while they run. out is written from other goroutines, so it
	// must be safe for that, and it must not block for long. The command
	// leads its own process group, and its process is an Interrupter.
	StartStreaming(ctx context.Context, out io.Writer, name string, args ...string) (Process, error)
}

// ExecRunner is the CommandRunner on top of tools/exec.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	p, err := startCommand(ctx, argv(name, args), captureOutput(&stdout, &stderr))
	if err != nil {
		return nil, err
	}
	if err := p.Wait(); err != nil {
		return stdout.Bytes(), &CommandError{Name: name, Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Err: err}
	}
	return stdout.Bytes(), nil
}

func (ExecRunner) Start(ctx context.Context, name string, args ...string) (Process, error) {
	return process(startCommand(ctx, argv(name, args)))
}

//nolint:revive // argument-limit: fixed by device.CommandRunner
func (ExecRunner) StartStreaming(ctx context.Context, out io.Writer, name string, args ...string) (Process, error) {
	return process(startCommand(ctx, argv(name, args), streamOutput(out)))
}

// process keeps a failed start from becoming a non-nil Process.
func process(p *execProcess, err error) (Process, error) {
	if err != nil {
		return nil, err
	}
	return p, nil
}

func argv(name string, args []string) []string { return append([]string{name}, args...) }
