package iosdevice

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

type fakeSource struct {
	frames [][]byte
	block  chan struct{}
	closed bool
}

func (f *fakeSource) ReadFrame() ([]byte, error) {
	if f.block != nil {
		<-f.block
		return nil, io.EOF
	}
	if len(f.frames) == 0 {
		return nil, io.EOF
	}
	frame := f.frames[0]
	f.frames = f.frames[1:]
	return frame, nil
}

func (f *fakeSource) Close() error {
	if !f.closed && f.block != nil {
		close(f.block)
	}
	f.closed = true
	return nil
}

func jpegFrame(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, w, h)), nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestVideoStreamSendsTheFirstFrameAndKeyFrames(t *testing.T) {
	first, second := jpegFrame(t, 30, 64), jpegFrame(t, 64, 30)
	src := &fakeSource{frames: [][]byte{first, second}}
	s, err := newVideoStream(src, device.VideoMeta{Codec: device.CodecMJPEG, ViewOnly: true, ScreenWidth: 1170, ScreenHeight: 2532}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	meta := s.Meta()
	if meta.Width != 30 || meta.Height != 64 || meta.Codec != device.CodecMJPEG || !meta.ViewOnly || meta.ScreenWidth != 1170 {
		t.Fatalf("meta = %+v", meta)
	}
	for i, want := range [][]byte{first, second} {
		p, err := s.ReadPacket()
		if err != nil {
			t.Fatalf("packet %d: %v", i, err)
		}
		if !p.KeyFrame || p.Config || !bytes.Equal(p.Data, want) {
			t.Fatalf("packet %d = %+v", i, p)
		}
	}
	if _, err := s.ReadPacket(); !errors.Is(err, io.EOF) {
		t.Fatalf("after the last frame: %v", err)
	}
	if _, ok := any(s).(device.VideoInput); ok {
		t.Fatal("the MJPEG stream must not claim live input")
	}
}

func TestVideoStreamFailsWithoutFrames(t *testing.T) {
	src := &fakeSource{block: make(chan struct{})}
	if _, err := newVideoStream(src, device.VideoMeta{}, 50*time.Millisecond); err == nil {
		t.Fatal("expected a timeout")
	}
	if !src.closed {
		t.Fatal("the source must be closed after a timeout")
	}
	if _, err := newVideoStream(&fakeSource{frames: [][]byte{[]byte("no jpeg")}}, device.VideoMeta{}, time.Second); err == nil {
		t.Fatal("expected a decode error")
	}
}

func TestMjpegScaling(t *testing.T) {
	iphone := webdriver.ScreenInfo{Width: 390, Height: 844, Scale: 3} // 1170 x 2532
	for _, tc := range []struct {
		max  int
		want int
	}{{0, 100}, {4096, 100}, {2532, 100}, {1280, 50}, {640, 25}, {1, 1}} {
		if got := mjpegScaling(iphone, tc.max); got != tc.want {
			t.Errorf("mjpegScaling(%d) = %d, want %d", tc.max, got, tc.want)
		}
	}
	landscape := webdriver.ScreenInfo{Width: 844, Height: 390, Scale: 3}
	if w, h := screenPixels(landscape); w != 1170 || h != 2532 {
		t.Errorf("screenPixels = %dx%d", w, h)
	}
}
