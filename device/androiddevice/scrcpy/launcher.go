package scrcpy

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/adb"
)

// Deps are what a Launcher needs from the outside.
type Deps struct {
	Runner device.CommandRunner
	Clock  device.Clock
	// SessionID makes the socket and jar names of a session unique; nil
	// uses random ids.
	SessionID func() int32
}

// Launcher starts scrcpy sessions. It writes the embedded server to disk once.
type Launcher struct {
	deps Deps
	jar  jarFile
}

// NewLauncher creates a launcher.
func NewLauncher(deps Deps) *Launcher {
	if deps.SessionID == nil {
		deps.SessionID = randomSessionID
	}
	return &Launcher{deps: deps}
}

func randomSessionID() int32 { return rand.Int32N(0x7fffffff) + 1 }

// deployment is the server on one device: its id, jar and forwarded port.
type deployment struct {
	scid   int32
	socket string
	remote string
	port   int
}

// Start pushes and starts the server on the device and connects to its video
// socket. The session ends when ctx is done or Close is called.
func (l *Launcher) Start(ctx context.Context, serial string, opts Options) (*Session, error) {
	c := adb.Client{Runner: l.deps.Runner, Serial: serial}
	dep, err := l.deploy(ctx, c)
	if err != nil {
		return nil, err
	}
	sctx, cancel := context.WithCancel(ctx)
	s := &Session{adb: c, clock: l.deps.Clock, port: dep.port, cancel: cancel, exited: make(chan struct{})}
	if err := s.launch(sctx, dep, opts); err != nil {
		cancel()
		return nil, err
	}
	if err := s.connect(sctx, opts.Control); err != nil {
		_ = s.Close()
		removeRemote(c, dep.remote)
		return nil, err
	}
	// close the socket when the caller's context ends, which unblocks reads
	go func() {
		<-sctx.Done()
		_ = s.Close()
	}()
	return s, nil
}

// deploy pushes the server jar and forwards a local port to its socket.
func (l *Launcher) deploy(ctx context.Context, c adb.Client) (deployment, error) {
	local, err := l.jar.local()
	if err != nil {
		return deployment{}, err
	}
	scid := l.deps.SessionID()
	dep := deployment{
		scid:   scid,
		socket: fmt.Sprintf("scrcpy_%08x", scid),
		remote: fmt.Sprintf("/data/local/tmp/automationhub-scrcpy-%08x.jar", scid),
	}
	if out, err := c.Run(ctx, "push", local, dep.remote); err != nil {
		return deployment{}, fmt.Errorf("push scrcpy server: %w: %s", err, out)
	}
	dep.port, err = forward(ctx, c, dep)
	return dep, err
}

// forward asks adb for a free local port ("tcp:0") that reaches the socket.
func forward(ctx context.Context, c adb.Client, dep deployment) (int, error) {
	out, err := c.Run(ctx, "forward", "tcp:0", "localabstract:"+dep.socket)
	if err != nil {
		removeRemote(c, dep.remote)
		return 0, fmt.Errorf("adb forward: %w: %s", err, out)
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		removeRemote(c, dep.remote)
		return 0, fmt.Errorf("adb forward returned %q", out)
	}
	return port, nil
}

// launch starts the server process; it ends with ctx.
func (s *Session) launch(ctx context.Context, dep deployment, opts Options) error {
	args := append([]string{"shell", "CLASSPATH=" + dep.remote, "app_process", "/", "com.genymobile.scrcpy.Server", Version},
		serverArgs(dep.scid, opts)...)
	s.output = &tailWriter{}
	proc, err := s.adb.StartStreaming(ctx, s.output, args...)
	if err != nil {
		s.removeForward()
		removeRemote(s.adb, dep.remote)
		return fmt.Errorf("start scrcpy server: %w", err)
	}
	go func() {
		s.exitErr = proc.Wait()
		close(s.exited)
	}()
	return nil
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
	return append(args, optionArgs(opts)...)
}

func optionArgs(opts Options) []string {
	var args []string
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

// removeRemote deletes the pushed jar when the server never started; a
// running server deletes it itself (cleanup=true).
func removeRemote(c adb.Client, remote string) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	_, _ = c.Run(ctx, "shell", "rm", "-f", remote)
}
