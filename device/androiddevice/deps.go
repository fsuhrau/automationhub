package androiddevice

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/tools/androidsdk"
	"github.com/google/uuid"
)

// Deps are the outside world of the Android devices. Nil fields use the
// defaults: the real adb, the system clock, random ids and the aapt of the
// Android SDK.
type Deps struct {
	Runner device.CommandRunner
	Clock  device.Clock
	// NewID names temporary files.
	NewID func() string
	// AAPT finds the aapt executable.
	AAPT func() (string, error)
	// SessionID names the scrcpy sockets, see scrcpy.Deps.
	SessionID func() int32
}

func (d Deps) withDefaults() Deps {
	if d.Runner == nil {
		d.Runner = device.ExecRunner{}
	}
	if d.Clock == nil {
		d.Clock = device.RealClock{}
	}
	if d.NewID == nil {
		d.NewID = uuid.NewString
	}
	if d.AAPT == nil {
		d.AAPT = androidsdk.AAPT
	}
	return d
}
