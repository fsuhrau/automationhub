package unityeditor

import (
	"runtime"

	"github.com/fsuhrau/automationhub/device"
)

func (d Deps) withDefaults() Deps {
	if d.Runner == nil {
		d.Runner = device.ExecRunner{}
	}
	if d.Clock == nil {
		d.Clock = device.RealClock{}
	}
	if d.GOOS == "" {
		d.GOOS = runtime.GOOS
	}
	return d
}
