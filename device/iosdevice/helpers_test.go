package iosdevice

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// waitFor polls cond until it holds; the conditions are goroutines catching up,
// not elapsed time.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		time.Sleep(time.Millisecond)
	}
}

// waitPending waits until n waiters sit on the fake clock.
func waitPending(t *testing.T, c *testkit.FakeClock, n int) {
	t.Helper()
	waitFor(t, func() bool { return c.Pending() >= n })
}

type deviceEntry = ios.DeviceEntry

// recordingWriter is a device.LogWriter that keeps the errors.
type recordingWriter struct {
	device.LogWriter
	errors []string
}

func (w *recordingWriter) Log(string, string, ...interface{}) {}
func (w *recordingWriter) Error(_, format string, args ...interface{}) {
	w.errors = append(w.errors, fmt.Sprintf(format, args...))
}
