package testkit

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/fsuhrau/automationhub/device"
)

// Process is the handle returned by FakeRunner.Start, the device.Process.
type Process = device.Process

type result struct {
	out    []byte
	stderr []byte
	err    error
}

// FakeRunner is a scripted command runner with the signatures of
// device.CommandRunner. Commands are keyed by their command line, the name and
// the arguments joined by single spaces.
type FakeRunner struct {
	mu      sync.Mutex
	scripts map[string]result
	lines   []string
}

// Script sets the output and error of a command line, for Run and Start.
func (r *FakeRunner) Script(line string, out []byte, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.scripts == nil {
		r.scripts = map[string]result{}
	}
	r.scripts[line] = result{out: out, err: err}
}

// ScriptStderr sets what the command line writes to stderr. Run then fails
// with a *device.CommandError that carries it, as ExecRunner does, instead of
// the plain scripted error. StartStreaming writes it to out after the output.
func (r *FakeRunner) ScriptStderr(line string, stderr []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.scripts == nil {
		r.scripts = map[string]result{}
	}
	res := r.scripts[line]
	res.stderr = stderr
	r.scripts[line] = res
}

// Commands returns the command lines of all Run and Start calls in order.
func (r *FakeRunner) Commands() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

func (r *FakeRunner) lookup(name string, args []string) (result, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, line)
	res, ok := r.scripts[line]
	if !ok {
		return result{}, fmt.Errorf("testkit: unscripted command %q", line)
	}
	return res, nil
}

// Run returns the scripted output and error, or an error for an unscripted
// command line.
func (r *FakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	res, err := r.lookup(name, args)
	if err != nil {
		return nil, err
	}
	if res.err != nil && res.stderr != nil {
		return res.out, &device.CommandError{Name: name, Stdout: res.out, Stderr: res.stderr, Err: res.err}
	}
	return res.out, res.err
}

// Start returns a process whose Wait gives the scripted error. Its Kill is
// recorded in Killed. An unscripted command line is an error.
func (r *FakeRunner) Start(_ context.Context, name string, args ...string) (Process, error) {
	res, err := r.lookup(name, args)
	if err != nil {
		return nil, err
	}
	return &FakeProcess{waitErr: res.err}, nil
}

// StartStreaming writes the scripted output and stderr to out in the
// background. The returned process waits for that, then gives the scripted
// error. An unscripted command line is
// an error.
//
//nolint:revive // argument-limit: fixed by device.CommandRunner
func (r *FakeRunner) StartStreaming(_ context.Context, out io.Writer, name string, args ...string) (Process, error) {
	res, err := r.lookup(name, args)
	if err != nil {
		return nil, err
	}
	written := make(chan struct{})
	go func() { // like exec: out may be a pipe that is read after this call
		defer close(written)
		_, _ = out.Write(res.out)
		_, _ = out.Write(res.stderr)
	}()
	return &FakeProcess{waitErr: res.err, written: written}, nil
}

// FakeProcess is the Process of FakeRunner.Start.
type FakeProcess struct {
	mu      sync.Mutex
	waitErr error
	written chan struct{} // closed when the streamed output is written; may be nil
	killed  int
}

// Wait returns the scripted error, after the streamed output is written.
func (p *FakeProcess) Wait() error {
	if p.written != nil {
		<-p.written
	}
	return p.waitErr
}

// Kill counts the call.
func (p *FakeProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.killed++
	return nil
}

// Killed is the number of Kill calls.
func (p *FakeProcess) Killed() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.killed
}
