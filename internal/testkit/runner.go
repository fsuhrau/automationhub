package testkit

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Process is the handle returned by FakeRunner.Start. G12 adopts it as the
// result of device.CommandRunner.Start(ctx, name, args...) (Process, error).
type Process interface {
	// Wait blocks until the process ended and returns its error.
	Wait() error
	// Kill stops the process.
	Kill() error
}

type result struct {
	out []byte
	err error
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

// FakeProcess is the Process of FakeRunner.Start.
type FakeProcess struct {
	mu      sync.Mutex
	waitErr error
	killed  int
}

// Wait returns the scripted error.
func (p *FakeProcess) Wait() error { return p.waitErr }

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
