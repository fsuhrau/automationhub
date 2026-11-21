package generic

import (
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// recordingWriter is a log writer that records what it is given.
type recordingWriter struct {
	calls []string
	id    uint
}

func (w *recordingWriter) note(format string, args ...any) {
	w.calls = append(w.calls, fmt.Sprintf(format, args...))
}

func (w *recordingWriter) LogPerformance(checkpoint string, cpu, _, _, _, _ float64, _ string) {
	w.note("performance %s %v", checkpoint, cpu)
}
func (w *recordingWriter) Data(source, data string) { w.note("data %s %s", source, data) }
func (w *recordingWriter) SetKey(key, value string) { w.note("key %s %s", key, value) }
func (w *recordingWriter) Passed(passed bool)       { w.note("passed %v", passed) }
func (w *recordingWriter) TestProtocolId() *uint    { return &w.id }
func (w *recordingWriter) Device() interface{}      { return nil }
func (w *recordingWriter) Parent() device.LogWriter { return nil }
func (w *recordingWriter) HasPassed() bool          { return false }
func (w *recordingWriter) Log(source, format string, params ...interface{}) {
	w.note("log %s "+format, append([]any{source}, params...)...)
}
func (w *recordingWriter) Error(source, format string, params ...interface{}) {
	w.note("error %s "+format, append([]any{source}, params...)...)
}

func TestDeviceForwardsItsLogToTheWriter(t *testing.T) {
	w := &recordingWriter{id: 7}
	var d Device
	d.SetLogWriter(w)

	d.LogPerformance("start", 1, 2, 3, 4, 5, "")
	d.Data("screen", "shot.png")
	d.SetKey("player", "1")
	d.Log("step", "tap %d", 3)
	d.Passed(true)
	d.Error("step", "failed %s", "badly")
	d.Exception("step", "crashed %s", "hard")
	d.SetTestResult(models.TestResultSuccess)

	want := []string{
		"performance start 1", "data screen shot.png", "key player 1", "log step tap 3",
		"passed true", "error step failed badly", "error step crashed hard",
	}
	if fmt.Sprint(w.calls) != fmt.Sprint(want) {
		t.Fatalf("calls %v, want %v", w.calls, want)
	}
	if d.GetLogWriter() != device.LogWriter(w) || *d.TestProtocolId() != 7 {
		t.Fatalf("writer %v, protocol %v", d.GetLogWriter(), d.TestProtocolId())
	}
}

// Outside a test there is no protocol: the device logs to the standard
// logger, and what only a protocol can keep is dropped.
func TestDeviceWithoutAWriterLogsToTheStandardLogger(t *testing.T) {
	hook := logtest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })
	var d Device

	d.LogPerformance("start", 1, 2, 3, 4, 5, "")
	d.Data("screen", "shot.png")
	d.SetKey("player", "1")
	d.Log("step", "tap %d", 3)
	d.Passed(true)
	d.Error("step", "failed %s", "badly")
	d.Exception("step", "crashed %s", "hard")

	var messages []string
	for _, e := range hook.AllEntries() {
		messages = append(messages, e.Message)
	}
	want := []string{"run key player = 1 outside a test, dropped", "tap 3", "Test: true", "failed badly", "crashed hard"}
	if fmt.Sprint(messages) != fmt.Sprint(want) {
		t.Fatalf("logged %v, want %v", messages, want)
	}
	if d.TestProtocolId() != nil || d.GetLogWriter() != nil {
		t.Fatal("a device without a writer has no protocol")
	}
}
