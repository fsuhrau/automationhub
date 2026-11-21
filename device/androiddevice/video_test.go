package androiddevice

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/scrcpy"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestParseWmSize(t *testing.T) {
	cases := []struct {
		out  string
		w, h int
	}{
		{"Physical size: 1080x2400\n", 1080, 2400},
		{"Physical size: 1440x3200\nOverride size: 1080x2400\n", 1080, 2400},
		{"error: no devices", 0, 0},
	}
	for _, c := range cases {
		if w, h := parseWmSize(c.out); w != c.w || h != c.h {
			t.Errorf("parseWmSize(%q) = %dx%d, want %dx%d", c.out, w, h, c.w, c.h)
		}
	}
}

func TestKeyCode(t *testing.T) {
	if c, err := keyCode("BACK"); err != nil || c != KEYCODE_BACK {
		t.Fatalf("back: %d %v", c, err)
	}
	if c, err := keyCode("66"); err != nil || c != 66 {
		t.Fatalf("numeric: %d %v", c, err)
	}
	if _, err := keyCode("nope"); err == nil {
		t.Fatal("expected an error for an unknown key")
	}
}

// fakeSession records the injected input.
type fakeSession struct {
	calls  []string
	failOn string
	closed bool
}

func (f *fakeSession) record(call string) error {
	f.calls = append(f.calls, call)
	if call == f.failOn {
		return errBoom
	}
	return nil
}

func (f *fakeSession) InjectTouch(action int, pointerID int64, pos scrcpy.Position) error {
	return f.record("touch " + strconv.Itoa(action) + " " + strconv.FormatInt(pointerID, 10) + " " + strconv.Itoa(pos.X) + "," + strconv.Itoa(pos.Y) + " " + strconv.Itoa(pos.Width) + "x" + strconv.Itoa(pos.Height))
}
func (f *fakeSession) InjectKeycode(action, keycode int) error {
	return f.record("key " + strconv.Itoa(action) + " " + strconv.Itoa(keycode))
}
func (f *fakeSession) InjectText(text string) error { return f.record("text " + text) }
func (f *fakeSession) InjectScroll(pos scrcpy.Position, h, v float64) error {
	return f.record("scroll " + strconv.Itoa(pos.X) + " " + strconv.FormatFloat(h, 'f', 0, 64) + " " + strconv.FormatFloat(v, 'f', 0, 64))
}
func (f *fakeSession) ResetVideo() error                       { return f.record("reset") }
func (f *fakeSession) Codec() string                           { return "h264" }
func (f *fakeSession) Size() (int, int)                        { return 576, 1280 }
func (f *fakeSession) ReadPacket() (device.VideoPacket, error) { return device.VideoPacket{}, nil }
func (f *fakeSession) Close() error                            { f.closed = true; return nil }

func TestInjectInput(t *testing.T) {
	tests := []struct {
		name    string
		ev      device.InputEvent
		want    []string
		wantErr string
	}{
		{"touch down", device.InputEvent{Kind: device.InputTouch, Action: device.InputDown, PointerID: 3, X: 1, Y: 2, Width: 10, Height: 20}, []string{"touch 0 3 1,2 10x20"}, ""},
		{"touch move", device.InputEvent{Kind: device.InputTouch, Action: device.InputMove, Width: 1, Height: 1}, []string{"touch 2 0 0,0 1x1"}, ""},
		{"touch up", device.InputEvent{Kind: device.InputTouch, Action: device.InputUp, Width: 1, Height: 1}, []string{"touch 1 0 0,0 1x1"}, ""},
		{"touch cancel", device.InputEvent{Kind: device.InputTouch, Action: device.InputCancel, Width: 1, Height: 1}, []string{"touch 3 0 0,0 1x1"}, ""},
		{"touch unknown action", device.InputEvent{Kind: device.InputTouch, Action: "x", Width: 1, Height: 1}, nil, `unknown touch action "x"`},
		{"touch without size", device.InputEvent{Kind: device.InputTouch, Action: device.InputDown}, nil, "touch needs the video width and height"},
		{"key down", device.InputEvent{Kind: device.InputKey, Key: "back", Action: device.InputDown}, []string{"key 0 4"}, ""},
		{"key up", device.InputEvent{Kind: device.InputKey, Key: "back", Action: device.InputUp}, []string{"key 1 4"}, ""},
		{"key press", device.InputEvent{Kind: device.InputKey, Key: "back", Action: device.InputPress}, []string{"key 0 4", "key 1 4"}, ""},
		{"key default is press", device.InputEvent{Kind: device.InputKey, Key: "66"}, []string{"key 0 66", "key 1 66"}, ""},
		{"key unknown", device.InputEvent{Kind: device.InputKey, Key: "nope"}, nil, `unknown key "nope"`},
		{"key unknown action", device.InputEvent{Kind: device.InputKey, Key: "back", Action: "x"}, nil, `unknown key action "x"`},
		{"text", device.InputEvent{Kind: device.InputText, Text: "hi"}, []string{"text hi"}, ""},
		{"scroll", device.InputEvent{Kind: device.InputScroll, X: 5, ScrollX: 1, ScrollY: -2}, []string{"scroll 5 1 -2"}, ""},
		{"reset video", device.InputEvent{Kind: device.InputResetVideo}, []string{"reset"}, ""},
		{"unknown kind", device.InputEvent{Kind: "x"}, nil, `unknown input kind "x"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeSession{}
			err := (&videoStream{videoSession: f}).InjectInput(tt.ev)
			if (err == nil) != (tt.wantErr == "") || (err != nil && !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if strings.Join(f.calls, "|") != strings.Join(tt.want, "|") {
				t.Fatalf("calls = %q, want %q", f.calls, tt.want)
			}
		})
	}
}

func TestPressKeyStopsAtTheFirstFailure(t *testing.T) {
	f := &fakeSession{failOn: "key 0 4"}
	err := (&videoStream{videoSession: f}).InjectInput(device.InputEvent{Kind: device.InputKey, Key: "back"})
	if !errors.Is(err, errBoom) || len(f.calls) != 1 {
		t.Fatalf("err %v calls %q", err, f.calls)
	}
}

// fakeStarter hands out a fake session.
type fakeStarter struct {
	session *fakeSession
	err     error
	serial  string
	opts    scrcpy.Options
}

func (s *fakeStarter) Start(_ context.Context, serial string, opts scrcpy.Options) (videoSession, error) {
	s.serial, s.opts = serial, opts
	if s.err != nil {
		return nil, s.err
	}
	return s.session, nil
}

func TestStartVideo(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell CLASSPATH=/data/local/tmp/automationhub-scrcpy-00001234.jar app_process / com.genymobile.scrcpy.Server "+scrcpy.Version+
		" scid=00001234 log_level=info video=true audio=false control=true clipboard_autosync=false video_codec=h264 tunnel_forward=true cleanup=true"+
		" send_device_meta=false send_frame_meta=true send_codec_meta=true send_dummy_byte=true", nil, nil)
	runner.Script("adb -s S1 shell wm size", []byte("Physical size: 1080x2400\n"), nil)
	starter := &fakeStarter{session: &fakeSession{}}
	d := newTestDevice(runner)
	d.video = starter

	stream, err := d.StartVideo(context.Background(), device.VideoOptions{MaxSize: 1280, MaxFPS: 30, BitRate: 4000000})
	if err != nil {
		t.Fatal(err)
	}
	want := device.VideoMeta{Codec: "h264", Width: 576, Height: 1280, ScreenWidth: 1080, ScreenHeight: 2400}
	if stream.Meta() != want {
		t.Fatalf("meta = %+v, want %+v", stream.Meta(), want)
	}
	if starter.serial != "S1" || starter.opts != (scrcpy.Options{MaxSize: 1280, MaxFPS: 30, BitRate: 4000000, Control: true}) {
		t.Fatalf("started %q with %+v", starter.serial, starter.opts)
	}
	if _, ok := stream.(device.VideoInput); !ok {
		t.Fatal("the stream must take input")
	}
}

func TestStartVideoFailures(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{})
	d.video = &fakeStarter{err: errBoom}
	if _, err := d.StartVideo(context.Background(), device.VideoOptions{}); !errors.Is(err, errBoom) {
		t.Fatalf("starter error = %v", err)
	}
	d.State = device.StateUnauthorized
	if _, err := d.StartVideo(context.Background(), device.VideoOptions{}); err == nil || errors.Is(err, errBoom) {
		t.Fatalf("not ready error = %v", err)
	}
}

func TestStartVideoWithoutWmSizeKeepsTheVideoSize(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{})
	d.video = &fakeStarter{session: &fakeSession{}}
	stream, err := d.StartVideo(context.Background(), device.VideoOptions{})
	if err != nil || stream.Meta().ScreenWidth != 0 || stream.Meta().Width != 576 {
		t.Fatalf("meta %+v err %v", stream.Meta(), err)
	}
}

// hangingProcess is a scrcpy server process that runs until it is killed.
type hangingProcess struct {
	done chan struct{}
	once sync.Once
}

func (p *hangingProcess) Wait() error { <-p.done; return errBoom }
func (p *hangingProcess) Kill() error { p.once.Do(func() { close(p.done) }); return nil }

// serverRunner keeps started processes running until their context ends.
type serverRunner struct{ *testkit.FakeRunner }

func (r serverRunner) Start(ctx context.Context, name string, args ...string) (device.Process, error) {
	if _, err := r.FakeRunner.Start(ctx, name, args...); err != nil {
		return nil, err
	}
	p := &hangingProcess{done: make(chan struct{})}
	go func() { <-ctx.Done(); _ = p.Kill() }()
	return p, nil
}

// TestStartVideoWithTheRealLauncher runs the scrcpy launcher against a fake
// server socket and checks the adb commands in order.
func TestStartVideoWithTheRealLauncher(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	go serveScrcpy(ln)

	runner := &testkit.FakeRunner{}
	deps := testDeps(runner)
	deps.Runner = serverRunner{runner}
	h := NewHandlerWithDeps(&memoryStore{}, deps)
	d := newDevice(h.deps, h.video, "S1")
	d.State = device.StateBooted
	runner.Script("adb -s S1 forward tcp:0 localabstract:scrcpy_00001234", []byte(port+"\n"), nil)
	runner.Script("adb -s S1 shell CLASSPATH=/data/local/tmp/automationhub-scrcpy-00001234.jar app_process / com.genymobile.scrcpy.Server "+scrcpy.Version+
		" scid=00001234 log_level=info video=true audio=false control=true clipboard_autosync=false video_codec=h264 tunnel_forward=true cleanup=true"+
		" send_device_meta=false send_frame_meta=true send_codec_meta=true send_dummy_byte=true", nil, nil)
	runner.Script("adb -s S1 shell wm size", []byte("Physical size: 720x1600"), nil)
	runner.Script("adb -s S1 forward --remove tcp:"+port, nil, nil)
	jar := scrcpyJarPush(t, runner)

	stream, err := d.StartVideo(context.Background(), device.VideoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if m := stream.Meta(); m.Width != 576 || m.Height != 1280 || m.ScreenWidth != 720 {
		t.Fatalf("meta = %+v", m)
	}
	_ = stream.Close()
	got := runner.Commands()
	if len(got) != 5 || got[0] != jar || got[1] != "adb -s S1 forward tcp:0 localabstract:scrcpy_00001234" ||
		!strings.Contains(got[2], "app_process / com.genymobile.scrcpy.Server "+scrcpy.Version) || got[3] != "adb -s S1 shell wm size" {
		t.Fatalf("commands = %q", got)
	}
}

// scrcpyJarPush scripts the push of the embedded server and returns its line.
func scrcpyJarPush(t *testing.T, runner *testkit.FakeRunner) string {
	t.Helper()
	// the jar is written to TMPDIR/automationhub-scrcpy-server-v<version>.jar
	line := "adb -s S1 push " + tmpJar() + " /data/local/tmp/automationhub-scrcpy-00001234.jar"
	runner.Script(line, nil, nil)
	runner.Script("adb -s S1 shell rm -f /data/local/tmp/automationhub-scrcpy-00001234.jar", nil, nil)
	return line
}

func tmpJar() string {
	return strings.TrimRight(os.TempDir(), "/") + "/automationhub-scrcpy-server-v" + scrcpy.Version + ".jar"
}

// serveScrcpy plays the server: the dummy byte, then the h264 header with the
// size 576x1280 once the control socket connected.
func serveScrcpy(ln net.Listener) {
	video, err := ln.Accept()
	if err != nil {
		return
	}
	_, _ = video.Write([]byte{0})
	if _, err := ln.Accept(); err != nil {
		return
	}
	header := make([]byte, 12)
	binary.BigEndian.PutUint32(header, 0x68323634)
	binary.BigEndian.PutUint32(header[4:], 576)
	binary.BigEndian.PutUint32(header[8:], 1280)
	_, _ = video.Write(header)
}

func TestScrcpyStarterReportsALauncherFailure(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	h := NewHandlerWithDeps(&memoryStore{}, testDeps(&testkit.FakeRunner{}))
	if _, err := h.video.Start(context.Background(), "S1", scrcpy.Options{}); err == nil {
		t.Fatal("expected the push error")
	}
}
