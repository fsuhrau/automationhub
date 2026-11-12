package base

import (
	"sync"

	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/tester/recording"
)

// RecordSessions records the screens of the devices when the test asks for
// it. The returned func stops the recordings and waits until their files are
// complete; call it before the devices are unlocked.
func (tr *TestRunner) RecordSessions(devices []DeviceMap) func() {
	if !tr.Config.RecordSession {
		return func() {}
	}
	recorders := make([]*recording.Recorder, 0, len(devices))
	for _, d := range devices {
		recorders = append(recorders, recording.Start(d.Device, recording.Options{
			Dir:      apps.TestDataPath,
			DB:       tr.DB,
			RunID:    tr.TestRun.ID,
			DeviceID: d.Model.ID,
			Log:      tr,
		}))
	}
	return func() {
		var wg sync.WaitGroup
		for _, r := range recorders {
			wg.Add(1)
			go func(r *recording.Recorder) {
				defer wg.Done()
				r.Stop()
			}(r)
		}
		wg.Wait()
	}
}
