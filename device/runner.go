package device

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	toolsexec "github.com/fsuhrau/automationhub/tools/exec"
)

// Process is a started command.
type Process interface {
	Wait() error
	Kill() error
}

// CommandRunner runs external commands for the platform packages.
type CommandRunner interface {
	// Run runs the command and returns its stdout.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	// Start starts the command in the background; cancelling ctx kills it.
	Start(ctx context.Context, name string, args ...string) (Process, error)
}

// ExecRunner is the CommandRunner on top of tools/exec.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	var out bytes.Buffer
	p, err := startCommand(ctx, &out, argv(name, args))
	if err != nil {
		return nil, err
	}
	if err := p.Wait(); err != nil {
		return out.Bytes(), fmt.Errorf("run %s: %w", name, err)
	}
	return out.Bytes(), nil
}

func (ExecRunner) Start(ctx context.Context, name string, args ...string) (Process, error) {
	return startCommand(ctx, nil, argv(name, args))
}

type execProcess struct {
	cmd  *exec.Cmd
	done chan struct{}
	once sync.Once
}

// startCommand starts argv, the command name followed by its arguments.
func startCommand(ctx context.Context, stdout io.Writer, argv []string) (*execProcess, error) {
	name := argv[0]
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	cmd := toolsexec.NewCommand(name, argv[1:]...)
	cmd.Stdout = stdout
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	p := &execProcess{cmd: cmd, done: make(chan struct{})}
	go p.killOnCancel(ctx)
	return p, nil
}

func (p *execProcess) killOnCancel(ctx context.Context) {
	select {
	case <-ctx.Done():
		_ = p.Kill()
	case <-p.done:
	}
}

func (p *execProcess) Wait() error {
	err := p.cmd.Wait()
	p.once.Do(func() { close(p.done) })
	return err
}

func (p *execProcess) Kill() error { return p.cmd.Process.Kill() }

func argv(name string, args []string) []string { return append([]string{name}, args...) }
