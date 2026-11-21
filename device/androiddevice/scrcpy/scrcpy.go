// Package scrcpy streams the screen of an Android device as H.264 video with
// the scrcpy server (https://github.com/Genymobile/scrcpy, Apache-2.0).
//
// The server jar is embedded, pushed to the device per session and started
// with app_process. It encodes the display with MediaCodec and writes the
// packets to a local abstract socket, which is reached through "adb forward".
// Audio is disabled. With Options.Control the session also opens the control
// socket, which injects touch, key and text events without "adb shell input".
package scrcpy

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/adb"
	"github.com/sirupsen/logrus"
)

const (
	closeWait      = 3 * time.Second
	cleanupTimeout = 5 * time.Second
)

// Options of a session. Zero values use the scrcpy defaults.
type Options struct {
	MaxSize int
	MaxFPS  int
	BitRate int
	// Control opens the control channel for Inject* calls.
	Control bool
}

// Session is a running scrcpy server with its video socket. It implements the
// packet reading part of device.VideoStream.
type Session struct {
	adb    adb.Client
	clock  device.Clock
	port   int
	conn   net.Conn
	reader *bufio.Reader
	codec  string
	width  int
	height int

	control   net.Conn
	controlMu sync.Mutex

	cancel    context.CancelFunc
	exited    chan struct{}
	exitErr   error
	closeOnce sync.Once
}

// Codec returns the codec of the stream ("h264").
func (s *Session) Codec() string { return s.codec }

// Size returns the initial video size.
func (s *Session) Size() (int, int) { return s.width, s.height }

// ReadPacket reads the next packet.
func (s *Session) ReadPacket() (device.VideoPacket, error) {
	return readPacket(s.reader)
}

// Close stops the server and removes the port forward.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		if s.conn != nil {
			_ = s.conn.Close()
		}
		if s.control != nil {
			_ = s.control.Close()
		}
		s.cancel()
		s.waitForExit()
		s.removeForward()
	})
	return nil
}

func (s *Session) waitForExit() {
	select {
	case <-s.exited:
	case <-s.clock.After(closeWait):
		logrus.Warnf("scrcpy server on %s did not exit", s.adb.Serial)
	}
}

func (s *Session) removeForward() {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	if out, err := s.adb.Run(ctx, "forward", "--remove", "tcp:"+strconv.Itoa(s.port)); err != nil {
		logrus.Debugf("remove adb forward %d on %s: %v: %s", s.port, s.adb.Serial, err, out)
	}
}
