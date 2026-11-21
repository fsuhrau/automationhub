// Package hostinfo reads facts about the machine a node runs on (OS name and
// version, host name, hardware identity) through an injected command runner,
// so the macos and web platforms share one implementation that tests on any OS.
package hostinfo

import (
	"runtime"

	"github.com/fsuhrau/automationhub/device"
)

// Env is what the host-dependent platform code gets injected. The zero value
// of a field means the production default.
type Env struct {
	Runner device.CommandRunner
	Clock  device.Clock
	// GOOS is the operating system to act for; it defaults to runtime.GOOS.
	GOOS string
}

// WithDefaults returns the Env with every unset field filled in.
func (e Env) WithDefaults() Env {
	if e.Runner == nil {
		e.Runner = device.ExecRunner{}
	}
	if e.Clock == nil {
		e.Clock = device.RealClock{}
	}
	if e.GOOS == "" {
		e.GOOS = runtime.GOOS
	}
	return e
}
