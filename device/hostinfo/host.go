package hostinfo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fsuhrau/automationhub/device"
)

const darwin = "darwin"

// Host reads the facts of one machine. Command output is cached by command
// line, the values do not change while a node runs.
type Host struct {
	runner device.CommandRunner
	goos   string

	mu    sync.Mutex
	cache map[string]string
}

// New returns a Host for the environment.
func New(env Env) *Host {
	env = env.WithDefaults()
	return &Host{runner: env.Runner, goos: env.GOOS, cache: map[string]string{}}
}

// output runs the command once and returns its raw stdout.
func (h *Host) output(ctx context.Context, name string, args ...string) (string, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	h.mu.Lock()
	cached, ok := h.cache[line]
	h.mu.Unlock()
	if ok {
		return cached, nil
	}
	out, err := h.runner.Run(ctx, name, args...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", line, err)
	}
	h.mu.Lock()
	h.cache[line] = string(out)
	h.mu.Unlock()
	return string(out), nil
}

// macOutput is output for the commands that only exist on macOS.
func (h *Host) macOutput(ctx context.Context, name string, args ...string) (string, error) {
	if h.goos != darwin {
		return "", fmt.Errorf("%s on %s: %w", name, h.goos, device.ErrUnsupported)
	}
	return h.output(ctx, name, args...)
}

func (h *Host) macTrimmed(ctx context.Context, name string, args ...string) (string, error) {
	out, err := h.macOutput(ctx, name, args...)
	return strings.TrimSpace(out), err
}
