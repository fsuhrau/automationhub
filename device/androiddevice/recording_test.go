package androiddevice

import (
	"os"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	recordLine = "adb -s S1 shell screenrecord --verbose /data/local/tmp/automation_hub_record.mp4"
	pkillLine  = "adb -s S1 shell pkill -2 screenrecord"
	pullLine   = "adb -s S1 pull /data/local/tmp/automation_hub_record.mp4"
)

func TestRecordingRoundTrip(t *testing.T) {
	dir := inTempDir(t)
	writeFile(t, "automation_hub_record.mp4", []byte("video"))
	runner := &testkit.FakeRunner{}
	scriptAll(runner, recordLine, pkillLine, pullLine)
	d := newTestDevice(runner)

	if err := d.StartRecording("run1"); err != nil {
		t.Fatal(err)
	}
	if err := d.StopRecording(); err != nil {
		t.Fatal(err)
	}
	assertCommands(t, runner, recordLine, pkillLine, pullLine)
	if data, err := os.ReadFile(dir + "/run1.mp4"); err != nil || string(data) != "video" {
		t.Fatalf("recording = %q, %v", data, err)
	}
	if d.recording.proc != nil {
		t.Fatal("the process must be released")
	}
}

func TestStopRecordingLogsAFailedWait(t *testing.T) {
	inTempDir(t)
	writeFile(t, "automation_hub_record.mp4", []byte("video"))
	runner := &testkit.FakeRunner{}
	runner.Script(recordLine, nil, errBoom) // Wait of the adb session fails
	scriptAll(runner, pkillLine, pullLine)
	d := newTestDevice(runner)
	if err := d.StartRecording("run2"); err != nil {
		t.Fatal(err)
	}
	if err := d.StopRecording(); err != nil {
		t.Fatalf("a failing wait must only be logged: %v", err)
	}
}

func TestStopRecordingWithoutStart(t *testing.T) {
	inTempDir(t)
	writeFile(t, "automation_hub_record.mp4", []byte("video"))
	runner := &testkit.FakeRunner{}
	scriptAll(runner, pullLine)
	d := newTestDevice(runner)
	d.recording.path = "./late.mp4"
	if err := d.StopRecording(); err != nil {
		t.Fatal(err)
	}
	assertCommands(t, runner, pullLine)
}

func TestRecordingFailures(t *testing.T) {
	t.Run("start", func(t *testing.T) {
		if err := newTestDevice(&testkit.FakeRunner{}).StartRecording("x"); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("interrupt", func(t *testing.T) {
		runner := &testkit.FakeRunner{}
		scriptAll(runner, recordLine)
		d := newTestDevice(runner)
		_ = d.StartRecording("x")
		if err := d.StopRecording(); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("pull", func(t *testing.T) {
		runner := &testkit.FakeRunner{}
		scriptAll(runner, recordLine, pkillLine)
		d := newTestDevice(runner)
		_ = d.StartRecording("x")
		if err := d.StopRecording(); err == nil {
			t.Fatal("expected an error")
		}
	})
}
