// Package recording records the screens of the devices of a test run, so the
// run can be replayed next to its log.
package recording

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"gorm.io/gorm"
)

// Logger writes to the run log.
type Logger interface {
	LogInfo(format string, params ...interface{})
	LogError(format string, params ...interface{})
}

// Options of a recording.
type Options struct {
	// Deps are the clock and file system; the zero value uses the real ones.
	Deps
	// Dir the files are written to; they are served from there by name.
	Dir      string
	DB       *gorm.DB
	RunID    uint
	DeviceID uint
	Log      Logger
	// Video asks the device for a smaller stream than the live view does:
	// a recording is kept and downloaded later.
	Video device.VideoOptions
}

const (
	// restartDelay is the wait before the stream is opened again after it
	// ended, e.g. while the device restarts.
	restartDelay = 2 * time.Second
	// stopTimeout is how long Stop waits for the files to be complete.
	stopTimeout = 30 * time.Second
)

// defaultVideo keeps a recording at about 10 MB a minute.
func defaultVideo() device.VideoOptions {
	return device.VideoOptions{MaxSize: 1280, MaxFPS: 30, BitRate: 1500000}
}

// Recorder records the screen of one device until Stop.
type Recorder struct {
	cancel context.CancelFunc
	done   chan struct{}
	clock  Clock
}

// Start records the screen of the device in the background. Devices without
// a video stream are skipped with a note in the run log.
func Start(dev device.Device, o Options) *Recorder {
	ctx, cancel := context.WithCancel(context.Background())
	o.Deps = o.withDefaults()
	r := &Recorder{cancel: cancel, done: make(chan struct{}), clock: o.Clock}
	if o.Video == (device.VideoOptions{}) {
		o.Video = defaultVideo()
	}
	go func() {
		defer close(r.done)
		run(ctx, dev, o)
	}()
	return r
}

// Stop ends the recording and waits until its files are complete.
func (r *Recorder) Stop() {
	r.cancel()
	select {
	case <-r.done:
	case <-r.clock.After(stopTimeout):
	}
}
