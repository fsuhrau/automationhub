package scrcpy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	connectTimeout = 10 * time.Second
	dialTimeout    = time.Second
	dummyByteWait  = 2 * time.Second
	retryPause     = 100 * time.Millisecond
)

// connect dials the forwarded socket until the server accepts. adb accepts the
// TCP connection even before the server listens and then closes it, so the
// dummy byte tells whether the server is really connected. The server accepts
// the control socket before it writes the video header.
func (s *Session) connect(ctx context.Context, withControl bool) error {
	deadline := s.clock.Now().Add(connectTimeout)
	var lastErr error
	for s.clock.Now().Before(deadline) {
		if err := s.alive(ctx); err != nil {
			return err
		}
		retry, err := s.attempt(withControl)
		if !retry {
			return err
		}
		lastErr = err
		<-s.clock.After(retryPause)
	}
	return fmt.Errorf("connect to scrcpy server: %v", lastErr)
}

// alive fails once the caller gave up or the server process ended.
func (s *Session) alive(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.exited:
		return s.exitError()
	default:
		return nil
	}
}

// exitError says that the server ended, with the end of its output.
func (s *Session) exitError() error {
	if tail := s.output.String(); tail != "" {
		return fmt.Errorf("scrcpy server exited: %v: %s", s.exitErr, tail)
	}
	return fmt.Errorf("scrcpy server exited: %v", s.exitErr)
}

// attempt tries to connect once. retry is true when err is transient.
func (s *Session) attempt(withControl bool) (retry bool, err error) {
	conn, err := s.dial(dialTimeout)
	if err != nil {
		return true, err
	}
	reader := bufio.NewReaderSize(conn, 64<<10)
	_ = conn.SetReadDeadline(time.Now().Add(dummyByteWait))
	if _, err := reader.ReadByte(); err != nil {
		_ = conn.Close()
		return true, err
	}
	if err := s.handshake(conn, reader, withControl); err != nil {
		_ = conn.Close()
		return false, err
	}
	return false, nil
}

// handshake opens the control socket and reads the video header.
func (s *Session) handshake(conn net.Conn, reader *bufio.Reader, withControl bool) error {
	if withControl {
		if err := s.openControl(); err != nil {
			return err
		}
	}
	_ = conn.SetReadDeadline(time.Now().Add(connectTimeout))
	var err error
	s.codec, s.width, s.height, err = readHeader(reader)
	if err != nil {
		return err
	}
	_ = conn.SetReadDeadline(time.Time{})
	s.conn = conn
	s.reader = reader
	return nil
}

func (s *Session) openControl() error {
	control, err := s.dial(connectTimeout)
	if err != nil {
		return fmt.Errorf("connect scrcpy control socket: %w", err)
	}
	s.control = control
	// the server only sends device messages (clipboard, uhid), which aren't
	// used; drain them so it never blocks
	go func() { _, _ = io.Copy(io.Discard, control) }()
	return nil
}

func (s *Session) dial(timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port)), timeout)
}
