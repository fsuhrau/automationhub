package androiddevice

import (
	"context"
	"regexp"
	"strconv"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/scrcpy"
)

var (
	physicalSizeRegex = regexp.MustCompile(`Physical size:\s*(\d+)x(\d+)`)
	overrideSizeRegex = regexp.MustCompile(`Override size:\s*(\d+)x(\d+)`)
)

// videoSession is a running scrcpy session.
type videoSession interface {
	injector
	Codec() string
	Size() (int, int)
	ReadPacket() (device.VideoPacket, error)
	Close() error
}

// sessionStarter starts scrcpy sessions.
type sessionStarter interface {
	Start(ctx context.Context, serial string, opts scrcpy.Options) (videoSession, error)
}

// scrcpyStarter adapts the scrcpy launcher to sessionStarter.
type scrcpyStarter struct{ launcher *scrcpy.Launcher }

func (s scrcpyStarter) Start(ctx context.Context, serial string, opts scrcpy.Options) (videoSession, error) {
	session, err := s.launcher.Start(ctx, serial, opts)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// StartVideo implements device.VideoStreamer with the scrcpy server.
func (d *Device) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	if err := d.notReady(); err != nil {
		return nil, err
	}
	d.Log("device", "Start video stream (scrcpy %s)", scrcpy.Version)
	session, err := d.video.Start(ctx, d.DeviceID(), scrcpy.Options{MaxSize: opts.MaxSize, MaxFPS: opts.MaxFPS, BitRate: opts.BitRate, Control: true})
	if err != nil {
		return nil, err
	}
	w, h := session.Size()
	meta := device.VideoMeta{Codec: session.Codec(), Width: w, Height: h}
	if out, err := d.run("shell", "wm", "size"); err == nil {
		meta.ScreenWidth, meta.ScreenHeight = parseWmSize(string(out))
	}
	return &videoStream{videoSession: session, meta: meta}, nil
}

// parseWmSize parses "wm size"; an override size wins, since input
// coordinates follow it.
func parseWmSize(out string) (int, int) {
	for _, re := range []*regexp.Regexp{overrideSizeRegex, physicalSizeRegex} {
		if m := re.FindStringSubmatch(out); m != nil {
			w, _ := strconv.Atoi(m[1])
			h, _ := strconv.Atoi(m[2])
			return w, h
		}
	}
	return 0, 0
}
