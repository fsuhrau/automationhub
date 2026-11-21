package web

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestStartAndStopApp(t *testing.T) {
	e := newTestEnv("darwin")
	d := &Device{env: e.env, browserPath: chromeBin}
	d.ID = uuid + "/chrome"
	line := chromeBin + " " + pageURL + "?sessionId=s1&nodeURL=n:1&deviceId=" + uuid + "/chrome"
	e.runner.Script(line, nil, nil)

	if err := d.StartApp(nil, webParams(), "s1", "n:1"); err != nil {
		t.Fatal(err)
	}
	if got := e.runner.Commands(); len(got) != 1 || got[0] != line {
		t.Fatalf("commands %v, want %q", got, line)
	}
	proc, ok := d.process.(*testkit.FakeProcess)
	if !ok {
		t.Fatalf("process %T", d.process)
	}
	if err := d.StopApp(nil); err != nil || proc.Killed() != 1 || d.process != nil {
		t.Fatalf("StopApp: %v, killed %d", err, proc.Killed())
	}
	if err := d.StopApp(nil); err != nil {
		t.Fatalf("StopApp without a browser: %v", err)
	}
}

func TestStartAppFailure(t *testing.T) {
	d := &Device{env: newTestEnv("linux").env, browserPath: chromeBin}
	if err := d.StartApp(nil, webParams(), "s", "n"); err == nil {
		t.Fatal("an unscripted start must fail")
	}
}

type failingProcess struct{}

func (failingProcess) Wait() error { return nil }
func (failingProcess) Kill() error { return errors.New("already exited") }

func TestStopAppKillFailure(t *testing.T) {
	d := &Device{process: failingProcess{}}
	if err := d.StopApp(nil); err == nil || d.process != nil {
		t.Fatalf("expected the kill error and a cleared process, got %v", err)
	}
}
