package base

import (
	"errors"

	"github.com/fsuhrau/automationhub/utils/sync"
)

// checkStarted reports whether the run can go on after the app start.
func (r *Runner) checkStarted(connected []DeviceMap, err error) bool {
	if errors.Is(err, sync.ErrTimeout) {
		r.LogError("Timeout while stating app")
	}
	if len(connected) == 0 {
		r.FailRun("no device connected, the tests can't run")
		return false
	}
	return true
}
