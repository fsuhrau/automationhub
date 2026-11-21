package recording

import (
	"context"
	"errors"

	"github.com/fsuhrau/automationhub/device"
)

// supervisor opens the video stream of a device again and again, and records
// each stream until it ends.
type supervisor struct {
	streamer device.VideoStreamer
	name     string
	o        Options
	failures int
}

func run(ctx context.Context, dev device.Device, o Options) {
	streamer, ok := dev.(device.VideoStreamer)
	if !ok {
		o.Log.LogInfo("Session recording: device %s has no video stream, it is not recorded", dev.DeviceID())
		return
	}
	s := &supervisor{streamer: streamer, name: dev.DeviceID(), o: o}
	for ctx.Err() == nil {
		if s.attempt(ctx) {
			return
		}
		select {
		case <-ctx.Done():
		case <-o.Clock.After(restartDelay):
		}
	}
}

// attempt opens and records one stream; it returns true when the device can't
// be recorded at all.
func (s *supervisor) attempt(ctx context.Context) (giveUp bool) {
	stream, err := s.streamer.StartVideo(ctx, s.o.Video)
	if err != nil {
		return s.failed(ctx, err)
	}
	if s.failures > 0 || ctx.Err() == nil {
		s.o.Log.LogInfo("Session recording: recording device %s", s.name)
	}
	s.failures = 0
	record(ctx, stream, s.o)
	return false
}

func (s *supervisor) failed(ctx context.Context, err error) (giveUp bool) {
	if errors.Is(err, device.ErrUnsupported) {
		s.o.Log.LogInfo("Session recording: device %s can't be recorded: %v", s.name, err)
		return true
	}
	if s.failures == 0 && ctx.Err() == nil {
		s.o.Log.LogError("Session recording: unable to start the video of device %s: %v", s.name, err)
	}
	s.failures++
	return false
}
