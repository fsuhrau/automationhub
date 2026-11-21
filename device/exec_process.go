package device

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	toolsexec "github.com/fsuhrau/automationhub/tools/exec"
)

// Interrupter is implemented by processes that can be asked to end before
// they are killed. The process of ExecRunner implements it.
type Interrupter interface {
	// Interrupt asks the process, and its children if it leads a process
	// group, to end.
	Interrupt() error
}

// setup changes a command before it starts.
type setup func(*exec.Cmd)

// captureOutput collects stdout and stderr; stderr still goes to the log.
func captureOutput(stdout, stderr io.Writer) setup {
	return func(cmd *exec.Cmd) {
		cmd.Stdout = stdout
		cmd.Stderr = io.MultiWriter(cmd.Stderr, stderr)
	}
}

// streamOutput sends stdout and stderr to out and runs the command in its own
// process group, so stopping it also ends the children it starts.
func streamOutput(out io.Writer) setup {
	return func(cmd *exec.Cmd) {
		setProcessGroup(cmd)
		cmd.Stdout = out
		cmd.Stderr = out
	}
}

type execProcess struct {
	cmd  *exec.Cmd
	done chan struct{}
	once sync.Once
}

// startCommand starts argv, the command name followed by its arguments.
func startCommand(ctx context.Context, argv []string, opts ...setup) (*execProcess, error) {
	name := argv[0]
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	cmd := toolsexec.NewCommand(name, argv[1:]...)
	for _, opt := range opts {
		opt(cmd)
	}
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

func (p *execProcess) Kill() error      { return killProcess(p.cmd) }
func (p *execProcess) Interrupt() error { return interruptProcess(p.cmd) }
