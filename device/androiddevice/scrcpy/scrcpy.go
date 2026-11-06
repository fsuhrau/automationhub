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
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

// Version of the embedded server. The server refuses to start when the
// client passes a different version.
const Version = "3.3.4"

//go:embed scrcpy-server-v3.3.4
var serverJar []byte

const (
	connectTimeout = 10 * time.Second
	// maxPacketSize guards against a corrupt stream; real packets are far smaller
	maxPacketSize = 16 << 20

	codecH264 = 0x68323634 // "h264"
	codecH265 = 0x68323635 // "h265"
	codecAV1  = 0x00617631 // "av1"

	flagConfig   = uint64(1) << 63
	flagKeyFrame = uint64(1) << 62
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
	serial string
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
	output    *tailBuffer
	closeOnce sync.Once
}

// Start pushes and starts the server on the device and connects to its video
// socket. The session ends when ctx is done or Close is called.
func Start(ctx context.Context, serial string, opts Options) (*Session, error) {
	local, err := localJar()
	if err != nil {
		return nil, err
	}

	scid := rand.Int32N(0x7fffffff) + 1
	socket := fmt.Sprintf("scrcpy_%08x", scid)
	remote := fmt.Sprintf("/data/local/tmp/automationhub-scrcpy-%08x.jar", scid)

	if out, err := adb(ctx, serial, "push", local, remote); err != nil {
		return nil, fmt.Errorf("push scrcpy server: %w: %s", err, out)
	}

	out, err := adb(ctx, serial, "forward", "tcp:0", "localabstract:"+socket)
	if err != nil {
		removeRemote(serial, remote)
		return nil, fmt.Errorf("adb forward: %w: %s", err, out)
	}
	port, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		removeRemote(serial, remote)
		return nil, fmt.Errorf("adb forward returned %q", out)
	}

	sctx, cancel := context.WithCancel(ctx)
	s := &Session{
		serial: serial,
		port:   port,
		cancel: cancel,
		exited: make(chan struct{}),
		output: &tailBuffer{max: 4096},
	}

	args := append([]string{"-s", serial, "shell", "CLASSPATH=" + remote, "app_process", "/", "com.genymobile.scrcpy.Server", Version},
		serverArgs(scid, opts)...)
	cmd := exec.CommandContext(sctx, "adb", args...)
	cmd.Stdout = s.output
	cmd.Stderr = s.output
	if err := cmd.Start(); err != nil {
		cancel()
		s.removeForward()
		removeRemote(serial, remote)
		return nil, fmt.Errorf("start scrcpy server: %w", err)
	}
	go func() {
		_ = cmd.Wait()
		close(s.exited)
	}()

	if err := s.connect(sctx, opts.Control); err != nil {
		_ = s.Close()
		removeRemote(serial, remote)
		return nil, err
	}
	// close the socket when the caller's context ends, which unblocks reads
	go func() {
		<-sctx.Done()
		_ = s.Close()
	}()
	return s, nil
}

func serverArgs(scid int32, opts Options) []string {
	args := []string{
		fmt.Sprintf("scid=%08x", scid),
		"log_level=info",
		"video=true",
		"audio=false",
		"control=" + strconv.FormatBool(opts.Control),
		// don't send device clipboard changes over the control socket
		"clipboard_autosync=false",
		"video_codec=h264",
		"tunnel_forward=true",
		"cleanup=true",
		"send_device_meta=false",
		"send_frame_meta=true",
		"send_codec_meta=true",
		"send_dummy_byte=true",
	}
	if opts.MaxSize > 0 {
		args = append(args, "max_size="+strconv.Itoa(opts.MaxSize))
	}
	if opts.MaxFPS > 0 {
		args = append(args, "max_fps="+strconv.Itoa(opts.MaxFPS))
	}
	if opts.BitRate > 0 {
		args = append(args, "video_bit_rate="+strconv.Itoa(opts.BitRate))
	}
	return args
}

// connect dials the forwarded socket until the server accepts. adb accepts the
// TCP connection even before the server listens and then closes it, so the
// dummy byte tells whether the server is really connected. The server accepts
// the control socket before it writes the video header.
func (s *Session) connect(ctx context.Context, withControl bool) error {
	deadline := time.Now().Add(connectTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.exited:
			return fmt.Errorf("scrcpy server exited: %s", s.output.String())
		default:
		}

		conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port)), time.Second)
		if err == nil {
			reader := bufio.NewReaderSize(conn, 64<<10)
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err = reader.ReadByte(); err == nil {
				if withControl {
					control, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port)), connectTimeout)
					if err != nil {
						_ = conn.Close()
						return fmt.Errorf("connect scrcpy control socket: %w", err)
					}
					s.control = control
					// the server only sends device messages (clipboard, uhid),
					// which aren't used; drain them so it never blocks
					go func() { _, _ = io.Copy(io.Discard, control) }()
				}
				_ = conn.SetReadDeadline(time.Now().Add(connectTimeout))
				s.codec, s.width, s.height, err = readHeader(reader)
				if err != nil {
					_ = conn.Close()
					return err
				}
				_ = conn.SetReadDeadline(time.Time{})
				s.conn = conn
				s.reader = reader
				return nil
			}
			_ = conn.Close()
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("connect to scrcpy server: %v: %s", lastErr, s.output.String())
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
		select {
		case <-s.exited:
		case <-time.After(3 * time.Second):
			logrus.Warnf("scrcpy server on %s did not exit", s.serial)
		}
		s.removeForward()
	})
	return nil
}

func (s *Session) removeForward() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if out, err := adb(ctx, s.serial, "forward", "--remove", "tcp:"+strconv.Itoa(s.port)); err != nil {
		logrus.Debugf("remove adb forward %d on %s: %v: %s", s.port, s.serial, err, out)
	}
}

// removeRemote deletes the pushed jar when the server never started; a
// running server deletes it itself (cleanup=true).
func removeRemote(serial, remote string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = adb(ctx, serial, "shell", "rm", "-f", remote)
}

// readHeader reads the codec meta: codec id, width and height (big endian).
func readHeader(r io.Reader) (string, int, int, error) {
	var header [12]byte
	if _, err := io.ReadFull(r, header[:4]); err != nil {
		return "", 0, 0, fmt.Errorf("read scrcpy codec: %w", err)
	}
	var codec string
	switch id := binary.BigEndian.Uint32(header[:4]); id {
	case 0:
		return "", 0, 0, errors.New("scrcpy disabled the video stream")
	case 1:
		return "", 0, 0, errors.New("scrcpy video configuration error, see the device log")
	case codecH264:
		codec = "h264"
	case codecH265:
		codec = "h265"
	case codecAV1:
		codec = "av1"
	default:
		return "", 0, 0, fmt.Errorf("unknown scrcpy codec id %#x", id)
	}
	if _, err := io.ReadFull(r, header[4:]); err != nil {
		return "", 0, 0, fmt.Errorf("read scrcpy video size: %w", err)
	}
	return codec, int(binary.BigEndian.Uint32(header[4:8])), int(binary.BigEndian.Uint32(header[8:12])), nil
}

// readPacket reads one frame header (pts and flags, size) and its payload.
func readPacket(r io.Reader) (device.VideoPacket, error) {
	var header [12]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return device.VideoPacket{}, err
	}
	ptsFlags := binary.BigEndian.Uint64(header[:8])
	size := binary.BigEndian.Uint32(header[8:])
	if size == 0 || size > maxPacketSize {
		return device.VideoPacket{}, fmt.Errorf("invalid scrcpy packet size %d", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return device.VideoPacket{}, err
	}
	p := device.VideoPacket{Data: data}
	if ptsFlags&flagConfig != 0 {
		p.Config = true
	} else {
		p.KeyFrame = ptsFlags&flagKeyFrame != 0
		p.PTS = int64(ptsFlags &^ (flagConfig | flagKeyFrame))
	}
	return p, nil
}

var (
	jarOnce sync.Once
	jarPath string
	jarErr  error
)

// localJar writes the embedded server to a temp file once per process.
func localJar() (string, error) {
	jarOnce.Do(func() {
		path := filepath.Join(os.TempDir(), "automationhub-scrcpy-server-v"+Version+".jar")
		if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, serverJar) {
			jarPath = path
			return
		}
		tmp, err := os.CreateTemp(filepath.Dir(path), "scrcpy-server-*.jar")
		if err != nil {
			jarErr = err
			return
		}
		_, err = tmp.Write(serverJar)
		if cerr := tmp.Close(); err == nil {
			err = cerr
		}
		if err == nil {
			err = os.Rename(tmp.Name(), path)
		}
		if err != nil {
			_ = os.Remove(tmp.Name())
			jarErr = fmt.Errorf("write scrcpy server: %w", err)
			return
		}
		jarPath = path
	})
	return jarPath, jarErr
}

func adb(ctx context.Context, serial string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "adb", append([]string{"-s", serial}, args...)...).CombinedOutput()
	return string(out), err
}

// tailBuffer keeps the last max bytes of the server output for error messages.
type tailBuffer struct {
	mu  sync.Mutex
	max int
	buf []byte
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if len(b.buf) > b.max {
		b.buf = b.buf[len(b.buf)-b.max:]
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.TrimSpace(string(b.buf))
}
