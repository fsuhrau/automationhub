package base

import (
	"sync"

	"github.com/fsuhrau/automationhub/tester/recording"
)

// RecordSessions records the screens of the devices when the test asks for
// it. The returned func stops the recordings and waits until their files are
// complete; call it before the devices are unlocked.
func (r *Runner) RecordSessions(devices []DeviceMap) func() {
	if !r.Config.RecordSession {
		return func() {}
	}
	recorders := make([]*recording.Recorder, 0, len(devices))
	for _, d := range devices {
		recorders = append(recorders, recording.Start(d.Device, recording.Options{
			Dir:      r.DataDir,
			DB:       r.DB,
			RunID:    r.TestRun.ID,
			DeviceID: d.Model.ID,
			Log:      r,
		}))
	}
	return func() { stopRecorders(recorders) }
}

func stopRecorders(recorders []*recording.Recorder) {
	var wg sync.WaitGroup
	for _, rec := range recorders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec.Stop()
		}()
	}
	wg.Wait()
}
