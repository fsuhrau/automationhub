package scrcpy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// autoClock advances itself by the waited duration.
type autoClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *autoClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *autoClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	ch := make(chan time.Time, 1)
	ch <- c.now
	return ch
}

// blockingProcess is a server process that runs until it is killed.
type blockingProcess struct {
	done chan struct{}
	once sync.Once
}

func (p *blockingProcess) Wait() error { <-p.done; return errors.New("killed") }
func (p *blockingProcess) Kill() error { p.once.Do(func() { close(p.done) }); return nil }

// longRunner records the commands in a FakeRunner and keeps started processes
// running until they are killed or their context ends.
type longRunner struct{ *testkit.FakeRunner }

func (r longRunner) Start(ctx context.Context, name string, args ...string) (device.Process, error) {
	if _, err := r.FakeRunner.Start(ctx, name, args...); err != nil {
		return nil, err
	}
	p := &blockingProcess{done: make(chan struct{})}
	go func() { <-ctx.Done(); _ = p.Kill() }()
	return p, nil
}

func (r longRunner) StartStreaming(ctx context.Context, out io.Writer, name string, args ...string) (device.Process, error) {
	_, err := r.FakeRunner.StartStreaming(ctx, out, name, args...)
	if err != nil {
		return nil, err
	}
	p := &blockingProcess{done: make(chan struct{})}
	go func() { <-ctx.Done(); _ = p.Kill() }()
	return p, nil
}

func header264(w, h uint32) []byte { return header(codecH264, w, h) }

// fakeServer plays the scrcpy server on a loopback port.
type fakeServer struct {
	ln net.Listener
}

func listen(t *testing.T) *fakeServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return &fakeServer{ln: ln}
}

func (f *fakeServer) port() int { return f.ln.Addr().(*net.TCPAddr).Port }

// serve answers the first connection with the dummy byte and the header; with
// control it waits for the second connection before the header.
func (f *fakeServer) serve(control bool, hdr []byte) {
	go func() {
		video, err := f.ln.Accept()
		if err != nil {
			return
		}
		_, _ = video.Write([]byte{0})
		if control {
			if _, err := f.ln.Accept(); err != nil {
				return
			}
		}
		_, _ = video.Write(hdr)
	}()
}

func newSession(port int, clock device.Clock) *Session {
	return &Session{clock: clock, port: port, exited: make(chan struct{})}
}

func TestConnectNoControl(t *testing.T) {
	srv := listen(t)
	srv.serve(false, header264(1080, 2400))
	s := newSession(srv.port(), &autoClock{})
	if err := s.connect(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.conn.Close() }()
	if w, h := s.Size(); s.Codec() != "h264" || w != 1080 || h != 2400 || s.control != nil {
		t.Fatalf("codec %q size %dx%d control %v", s.Codec(), w, h, s.control)
	}
}

func TestConnectWithControlAndRetry(t *testing.T) {
	srv := listen(t)
	// the first connection is dropped like adb does before the server listens
	go func() {
		c, err := srv.ln.Accept()
		if err == nil {
			_ = c.Close()
		}
		srv.serve(true, header264(576, 1280))
	}()
	s := newSession(srv.port(), &autoClock{})
	if err := s.connect(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.conn.Close(); _ = s.control.Close() }()
	if s.control == nil || s.width != 576 {
		t.Fatalf("control %v width %d", s.control, s.width)
	}
}

func TestConnectFailures(t *testing.T) {
	closedPort := func() int {
		srv := listen(t)
		p := srv.port()
		_ = srv.ln.Close()
		return p
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name    string
		session func() *Session
		ctx     context.Context
		control bool
		want    string
	}{
		{"deadline", func() *Session { return newSession(closedPort(), &autoClock{}) }, context.Background(), false, "connect to scrcpy server"},
		{"cancelled", func() *Session { return newSession(closedPort(), &autoClock{}) }, cancelled, false, "context canceled"},
		{"server exited", func() *Session {
			s := newSession(closedPort(), &autoClock{})
			s.exitErr = errors.New("exit 1")
			close(s.exited)
			return s
		}, context.Background(), false, "scrcpy server exited: exit 1"},
		{"server exited with output", func() *Session {
			s := newSession(closedPort(), &autoClock{})
			s.exitErr = errors.New("exit 1")
			s.output = &tailWriter{}
			_, _ = s.output.Write([]byte("Aborted: no display\n"))
			close(s.exited)
			return s
		}, context.Background(), false, "scrcpy server exited: exit 1: Aborted: no display"},
		{"bad header", func() *Session {
			srv := listen(t)
			srv.serve(false, []byte{0, 0, 0, 1})
			return newSession(srv.port(), &autoClock{})
		}, context.Background(), false, "configuration error"},
		{"control socket refused", func() *Session {
			srv := listen(t)
			go func() {
				c, err := srv.ln.Accept()
				if err == nil {
					_, _ = c.Write([]byte{0})
					_ = srv.ln.Close()
				}
			}()
			return newSession(srv.port(), &autoClock{})
		}, context.Background(), true, "connect scrcpy control socket"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session().connect(tt.ctx, tt.control)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

// startFixture is a Launcher with a fake runner and a fake server.
type startFixture struct {
	runner  longRunner
	srv     *fakeServer
	l       *Launcher
	push    string
	forward string
}

func newStartFixture(t *testing.T) *startFixture {
	t.Helper()
	t.Setenv("TMPDIR", t.TempDir())
	runner := longRunner{&testkit.FakeRunner{}}
	l := NewLauncher(Deps{Runner: runner, Clock: &autoClock{}, SessionID: func() int32 { return 0x1234 }})
	jar, err := l.jar.local()
	if err != nil {
		t.Fatal(err)
	}
	f := &startFixture{runner: runner, srv: listen(t), l: l}
	f.push = "adb -s S1 push " + jar + " /data/local/tmp/automationhub-scrcpy-00001234.jar"
	f.forward = "adb -s S1 forward tcp:0 localabstract:scrcpy_00001234"
	runner.Script(f.push, nil, nil)
	runner.Script(f.forward, []byte(strconv.Itoa(f.srv.port())+"\n"), nil)
	runner.Script("adb -s S1 forward --remove tcp:"+strconv.Itoa(f.srv.port()), nil, nil)
	runner.Script("adb -s S1 shell rm -f /data/local/tmp/automationhub-scrcpy-00001234.jar", nil, nil)
	return f
}

func (f *startFixture) serverLine(opts Options) string {
	args := append([]string{"adb", "-s", "S1", "shell", "CLASSPATH=/data/local/tmp/automationhub-scrcpy-00001234.jar",
		"app_process", "/", "com.genymobile.scrcpy.Server", Version}, serverArgs(0x1234, opts)...)
	return strings.Join(args, " ")
}

func TestStartRunsPushForwardServerInOrder(t *testing.T) {
	f := newStartFixture(t)
	opts := Options{MaxSize: 1280, Control: true}
	f.runner.Script(f.serverLine(opts), nil, nil)
	f.srv.serve(true, header264(576, 1280))

	s, err := f.l.Start(context.Background(), "S1", opts)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := s.Size(); w != 576 || h != 1280 {
		t.Fatalf("size %dx%d", w, h)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	want := []string{f.push, f.forward, f.serverLine(opts), "adb -s S1 forward --remove tcp:" + strconv.Itoa(f.srv.port())}
	got := f.runner.Commands()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands\n got %q\nwant %q", got, want)
	}
}

func TestStartClosesWhenTheContextEnds(t *testing.T) {
	f := newStartFixture(t)
	f.runner.Script(f.serverLine(Options{}), nil, nil)
	f.srv.serve(false, header264(1, 1))
	ctx, cancel := context.WithCancel(context.Background())
	s, err := f.l.Start(ctx, "S1", Options{})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	<-s.exited
	if _, err := s.ReadPacket(); err == nil {
		t.Fatal("reading a closed session must fail")
	}
}

func TestStartFailures(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name  string
		setup func(f *startFixture)
		want  string
		rm    bool // the jar is removed from the device afterwards
	}{
		{"push", func(f *startFixture) { f.runner.Script(f.push, []byte("out"), boom) }, "push scrcpy server: " + boom.Error(), false},
		{"forward", func(f *startFixture) { f.runner.Script(f.forward, nil, boom) }, "adb forward: boom", true},
		{"forward port", func(f *startFixture) { f.runner.Script(f.forward, []byte("x"), nil) }, `adb forward returned "x"`, true},
		{"server start", nil, "start scrcpy server", true},
		{"connect", func(f *startFixture) {
			f.runner.Script(f.serverLine(Options{}), nil, nil)
			f.srv.serve(false, []byte{0, 0, 0, 1})
		}, "configuration error", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newStartFixture(t)
			if tt.setup != nil {
				tt.setup(f)
			}
			_, err := f.l.Start(context.Background(), "S1", Options{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			rm := "adb -s S1 shell rm -f /data/local/tmp/automationhub-scrcpy-00001234.jar"
			if got := strings.Contains(strings.Join(f.runner.Commands(), "\n"), rm); got != tt.rm {
				t.Fatalf("jar removed = %v, want %v: %q", got, tt.rm, f.runner.Commands())
			}
		})
	}
}

// The server output is the only place that says why scrcpy died.
func TestStartReportsTheOutputOfAServerThatDied(t *testing.T) {
	f := newStartFixture(t)
	_ = f.srv.ln.Close() // nothing to connect to, so the exit is what Start sees
	line := f.serverLine(Options{})
	f.runner.Script(line, []byte("starting server\n"), errors.New("exit status 1"))
	f.runner.ScriptStderr(line, []byte("java.lang.IllegalStateException: no display\n"))
	f.l.deps.Runner = f.runner.FakeRunner // the fake process ends at once

	_, err := f.l.Start(context.Background(), "S1", Options{})
	for _, want := range []string{"scrcpy server exited: exit status 1", "starting server", "java.lang.IllegalStateException: no display"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want it to contain %q", err, want)
		}
	}
}

func TestTailWriterKeepsTheEnd(t *testing.T) {
	var nilTail *tailWriter
	if nilTail.String() != "" {
		t.Error("a nil tail is empty")
	}
	tail := &tailWriter{}
	_, _ = tail.Write([]byte(strings.Repeat("a", outputTailSize)))
	_, _ = tail.Write([]byte("end\n"))
	got := tail.String()
	if !strings.HasSuffix(got, "aaend") || len(got) > outputTailSize {
		t.Errorf("tail = %d bytes ending %q", len(got), got[len(got)-5:])
	}
}

func TestStartFailsWithoutTheLocalJar(t *testing.T) {
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
	l := NewLauncher(Deps{Runner: &testkit.FakeRunner{}, Clock: &autoClock{}})
	if _, err := l.Start(context.Background(), "S1", Options{}); err == nil || !strings.Contains(err.Error(), "write scrcpy server") {
		t.Fatalf("error = %v", err)
	}
}

func TestWriteJar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	path, err := writeJar([]byte("jar"))
	if err != nil {
		t.Fatal(err)
	}
	// an identical file is reused
	before, _ := os.Stat(path)
	if again, err := writeJar([]byte("jar")); err != nil || again != path {
		t.Fatalf("reuse: %q %v", again, err)
	}
	if after, _ := os.Stat(path); !after.ModTime().Equal(before.ModTime()) {
		t.Fatal("identical jar was rewritten")
	}
	// a different one is replaced
	if _, err := writeJar([]byte("jar2")); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != "jar2" {
		t.Fatalf("jar = %q", data)
	}
}

func TestWriteJarFailures(t *testing.T) {
	t.Run("temp dir missing", func(t *testing.T) {
		t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
		if _, err := writeJar([]byte("x")); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("target is a directory", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("TMPDIR", dir)
		target := filepath.Join(dir, fmt.Sprintf("automationhub-scrcpy-server-v%s.jar", Version))
		if err := os.Mkdir(target, 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := writeJar([]byte("x")); err == nil {
			t.Fatal("expected an error")
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 1 {
			t.Fatalf("temp file left behind: %v", entries)
		}
	})
}

func TestCloseWarnsWhenTheServerDoesNotExit(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 forward --remove tcp:9", nil, errors.New("gone"))
	s := &Session{clock: &autoClock{}, port: 9, exited: make(chan struct{}), cancel: func() {}}
	s.adb.Runner = runner
	s.adb.Serial = "S1"
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if got := runner.Commands(); len(got) != 1 {
		t.Fatalf("commands = %v", got)
	}
}

func TestCloseReturnsOnceTheServerExited(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 forward --remove tcp:9", nil, nil)
	// the clock never fires, so only the exit can end the wait
	s := &Session{clock: testkit.NewFakeClock(time.Unix(0, 0)), port: 9, exited: make(chan struct{}), cancel: func() {}}
	s.adb.Runner = runner
	s.adb.Serial = "S1"
	close(s.exited)
	if err := s.Close(); err != nil || len(runner.Commands()) != 1 {
		t.Fatalf("err %v commands %v", err, runner.Commands())
	}
}
