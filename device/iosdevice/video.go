package iosdevice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"math"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

const (
	mjpegQuality = 60
	maxMjpegFPS  = 30
	// firstFrameTimeout bounds the wait for the first frame, which gives the
	// size of the video.
	firstFrameTimeout = 15 * time.Second
)

// mjpegSource is the part of webdriver.MjpegStream the video stream uses.
type mjpegSource interface {
	ReadFrame() ([]byte, error)
	Close() error
}

// videoStream turns the WebDriverAgent MJPEG stream into a device.VideoStream:
// every packet is a complete JPEG and a key frame. It has no input channel
// (meta.ViewOnly), so clients use the control API to tap and swipe.
type videoStream struct {
	src   mjpegSource
	meta  device.VideoMeta
	start time.Time

	mu    sync.Mutex
	first []byte // the frame read to learn the size, sent first
}

func (s *videoStream) Meta() device.VideoMeta { return s.meta }

func (s *videoStream) ReadPacket() (device.VideoPacket, error) {
	s.mu.Lock()
	frame := s.first
	s.first = nil
	s.mu.Unlock()
	if frame == nil {
		var err error
		if frame, err = s.src.ReadFrame(); err != nil {
			return device.VideoPacket{}, err
		}
	}
	return device.VideoPacket{KeyFrame: true, PTS: time.Since(s.start).Microseconds(), Data: frame}, nil
}

func (s *videoStream) Close() error { return s.src.Close() }

// mjpegScaling returns the mjpegScalingFactor (percent of the screen size)
// that keeps the longer side of the video within maxSize pixels.
func mjpegScaling(screen webdriver.ScreenInfo, maxSize int) int {
	longer := math.Max(screen.Width, screen.Height) * screen.Scale
	if maxSize <= 0 || longer <= 0 || float64(maxSize) >= longer {
		return 100
	}
	return max(1, int(math.Floor(float64(maxSize)*100/longer)))
}

// screenPixels is the screen size in pixels in portrait orientation, the
// coordinate space of Tap and Swipe in portrait.
func screenPixels(screen webdriver.ScreenInfo) (int, int) {
	w := int(math.Round(math.Min(screen.Width, screen.Height) * screen.Scale))
	h := int(math.Round(math.Max(screen.Width, screen.Height) * screen.Scale))
	return w, h
}

// StartVideo implements device.VideoStreamer with the MJPEG server of
// WebDriverAgent, which is started when it isn't running yet. The MJPEG
// settings apply to the whole runner, so the last viewer's options win.
func (d *Device) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	if state := d.DeviceState(); state != device.StateBooted {
		return nil, fmt.Errorf("device %s is %s", d.DeviceID(), device.StateToString(state))
	}
	wd, err := d.ensureWebDriver()
	if err != nil {
		return nil, fmt.Errorf("live video needs WebDriverAgent: %w", err)
	}
	address, err := wd.MjpegURL(webdriver.DefaultMjpegServerPort)
	if err != nil {
		return nil, err
	}
	d.Log("device", "Start video stream (WebDriverAgent MJPEG %s)", address)

	var meta device.VideoMeta
	meta.Codec = device.CodecMJPEG
	meta.ViewOnly = true
	fps := opts.MaxFPS
	if fps <= 0 || fps > maxMjpegFPS {
		fps = maxMjpegFPS
	}
	settings := map[string]any{
		"mjpegServerFramerate":         fps,
		"mjpegServerScreenshotQuality": mjpegQuality,
	}
	if screen, err := wd.Screen(); err == nil {
		meta.ScreenWidth, meta.ScreenHeight = screenPixels(screen)
		settings["mjpegScalingFactor"] = mjpegScaling(screen, opts.MaxSize)
	} else {
		d.Log("device", "screen size unknown: %v", err)
	}
	if err := wd.UpdateSettings(settings); err != nil {
		// the stream works with the runner's defaults as well
		d.Log("device", "MJPEG settings not applied: %v", err)
	}

	src, err := webdriver.OpenMjpeg(ctx, address)
	if err != nil {
		return nil, err
	}
	stream, err := newVideoStream(src, meta, firstFrameTimeout)
	if err != nil {
		_ = src.Close()
		return nil, err
	}
	return stream, nil
}

// newVideoStream reads the first frame to learn the video size.
func newVideoStream(src mjpegSource, meta device.VideoMeta, timeout time.Duration) (*videoStream, error) {
	type result struct {
		frame []byte
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		frame, err := src.ReadFrame()
		ch <- result{frame, err}
	}()
	var first []byte
	select {
	case r := <-ch:
		if r.err != nil {
			return nil, fmt.Errorf("read the first MJPEG frame: %w", r.err)
		}
		first = r.frame
	case <-time.After(timeout):
		_ = src.Close()
		return nil, errors.New("the MJPEG server sent no frame")
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(first))
	if err != nil {
		return nil, fmt.Errorf("decode the first MJPEG frame: %w", err)
	}
	meta.Width, meta.Height = cfg.Width, cfg.Height
	return &videoStream{src: src, meta: meta, start: time.Now(), first: first}, nil
}
