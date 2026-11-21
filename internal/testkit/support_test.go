package testkit

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

func TestRecorder(t *testing.T) {
	var r recorder
	r.FailWith("b", errBoom)
	if err := r.record("a", 1); err != nil {
		t.Errorf("a: %v", err)
	}
	if err := r.record("b"); err != errBoom {
		t.Errorf("b: %v", err)
	}
	_ = r.record("a", 2)
	if got := len(r.Calls()); got != 3 {
		t.Errorf("calls = %d", got)
	}
	if got := r.CallsTo("a"); !reflect.DeepEqual(got, []Call{{"a", []any{1}}, {"a", []any{2}}}) {
		t.Errorf("calls to a = %v", got)
	}
	r.FailWith("b", nil)
	if err := r.record("b"); err != nil {
		t.Errorf("cleared: %v", err)
	}
}

func TestFakeClockAfterFiresOnAdvance(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := NewFakeClock(start)
	short, long := c.After(time.Second), c.After(time.Minute)
	if c.Pending() != 2 {
		t.Fatalf("pending = %d", c.Pending())
	}
	c.Advance(2 * time.Second)
	select {
	case got := <-short:
		if !got.Equal(start.Add(2 * time.Second)) {
			t.Errorf("fired at %v", got)
		}
	default:
		t.Fatal("short timer did not fire")
	}
	select {
	case <-long:
		t.Fatal("long timer fired early")
	default:
	}
	if c.Pending() != 1 || !c.Now().Equal(start.Add(2*time.Second)) {
		t.Errorf("pending=%d now=%v", c.Pending(), c.Now())
	}
}

func TestFakeClockAfterZeroFiresAtOnce(t *testing.T) {
	c := NewFakeClock(time.Unix(5, 0))
	select {
	case <-c.After(0):
	default:
		t.Fatal("did not fire")
	}
}

func TestFakeRunnerRunIsScriptedAndRecorded(t *testing.T) {
	r := &FakeRunner{}
	r.Script("adb devices -l", []byte("out"), errBoom)
	out, err := r.Run(context.Background(), "adb", "devices", "-l")
	if string(out) != "out" || err != errBoom {
		t.Errorf("out=%q err=%v", out, err)
	}
	if _, err := r.Run(context.Background(), "adb"); err == nil {
		t.Error("unscripted command must fail")
	}
	if got := r.Commands(); !reflect.DeepEqual(got, []string{"adb devices -l", "adb"}) {
		t.Errorf("commands = %v", got)
	}
}

func TestFakeRunnerStart(t *testing.T) {
	r := &FakeRunner{}
	r.Script("scrcpy", nil, errBoom)
	p, err := r.Start(context.Background(), "scrcpy")
	if err != nil {
		t.Fatal(err)
	}
	if p.Wait() != errBoom {
		t.Error("wait error")
	}
	if err := p.Kill(); err != nil || p.(*FakeProcess).Killed() != 1 {
		t.Errorf("kill err=%v", err)
	}
	if _, err := r.Start(context.Background(), "other"); err == nil {
		t.Error("unscripted start must fail")
	}
}

func TestFakeRunnerRunWithStderrFailsLikeExecRunner(t *testing.T) {
	r := &FakeRunner{}
	r.Script("adb install x", []byte("out"), errBoom)
	r.ScriptStderr("adb install x", []byte("Failure [INSTALL_FAILED]\n"))
	out, err := r.Run(context.Background(), "adb", "install", "x")
	var cmdErr *device.CommandError
	if !errors.As(err, &cmdErr) || !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want a *device.CommandError wrapping errBoom", err)
	}
	if string(out) != "out" || string(cmdErr.Stdout) != "out" || string(cmdErr.Stderr) != "Failure [INSTALL_FAILED]\n" {
		t.Errorf("out=%q stdout=%q stderr=%q", out, cmdErr.Stdout, cmdErr.Stderr)
	}
}

func TestFakeRunnerStderrWithoutAnErrorIsNotAnError(t *testing.T) {
	r := &FakeRunner{}
	r.ScriptStderr("adb ok", []byte("warning"))
	r.Script("adb ok", []byte("fine"), nil)
	if out, err := r.Run(context.Background(), "adb", "ok"); err != nil || string(out) != "fine" {
		t.Errorf("out=%q err=%v", out, err)
	}
}

func TestFakeRunnerStartStreamingWritesOutputAndStderr(t *testing.T) {
	r := &FakeRunner{}
	r.Script("xcodebuild test", []byte("line1\n"), errBoom)
	r.ScriptStderr("xcodebuild test", []byte("line2\n"))
	pr, pw := io.Pipe()
	p, err := r.StartStreaming(context.Background(), pw, "xcodebuild", "test")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = p.Wait(); _ = pw.Close() }()
	got, _ := io.ReadAll(pr)
	if string(got) != "line1\nline2\n" {
		t.Errorf("streamed %q", got)
	}
	if p.Wait() != errBoom {
		t.Error("wait error")
	}
	if _, err := r.StartStreaming(context.Background(), io.Discard, "other"); err == nil {
		t.Error("unscripted streaming start must fail")
	}
	if got := r.Commands(); !reflect.DeepEqual(got, []string{"xcodebuild test", "other"}) {
		t.Errorf("commands = %v", got)
	}
}

func TestNewCoreDeviceHasOnlyTheCoreRoles(t *testing.T) {
	d := NewCoreDevice("c1")
	if d.DeviceID() != "c1" || d.Fake.DeviceID() != "c1" {
		t.Errorf("id = %q", d.DeviceID())
	}
	if _, ok := device.As[device.Screenshotter](d); ok {
		t.Error("a core device must not be a Screenshotter")
	}
	if err := d.InstallApp(nil); err != nil || len(d.Fake.CallsTo("InstallApp")) != 1 {
		t.Errorf("core calls reach the fake: %v", err)
	}
}

var (
	_ device.CommandRunner = (*FakeRunner)(nil)
	_ device.Clock         = (*FakeClock)(nil)
)

func TestRolesListsTheOptionalRolesOfADevice(t *testing.T) {
	full := []string{"Recorder", "Screenshotter", "FeatureExecutor", "NativeScripter", "AppDataClearer", "InputCapable", "LogReader", "LogFollower", "VideoStreamer"}
	if got := Roles(NewFakeDevice("a")); !reflect.DeepEqual(got, full) {
		t.Errorf("full device roles = %v", got)
	}
	if got := Roles(NewCoreDevice("a")); len(got) != 0 {
		t.Errorf("core device roles = %v", got)
	}
}
