package session

import (
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// flaky reports the app as connected from the connectAt-th check on.
type flaky struct {
	*testkit.FakeDevice
	checks, connectAt int
}

func (f *flaky) IsAppConnected() bool {
	f.checks++
	return f.checks >= f.connectAt
}

func sessionWith(clock Clock, dev *flaky) *Session {
	return &Session{clock: clock, Lock: &manager.DeviceLock{Device: dev}}
}

func newFlaky(timeout time.Duration, connectAt int) *flaky {
	d := testkit.NewFakeDevice("d1")
	d.Timeout = timeout
	return &flaky{FakeDevice: d, connectAt: connectAt}
}

// ticker advances the clock by one poll interval each time the waiter sleeps.
func tick(t *testing.T, clock *testkit.FakeClock, done <-chan error) error {
	t.Helper()
	for {
		select {
		case err := <-done:
			return err
		default:
		}
		if clock.Pending() > 0 {
			clock.Advance(time.Second)
		}
	}
}

func waitAsync(s *Session) <-chan error {
	done := make(chan error, 1)
	go func() { done <- s.WaitForConnection() }()
	return done
}

func TestWaitForConnection(t *testing.T) {
	cases := []struct {
		name      string
		connectAt int
		want      error
	}{
		{"connected at once", 1, nil},
		{"connects while polling", 3, nil},
		{"never connects", 1 << 30, ErrConnectionTimeout},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			clock := newClock()
			s := sessionWith(clock, newFlaky(3*time.Second, c.connectAt))
			if err := tick(t, clock, waitAsync(s)); !errors.Is(err, c.want) {
				t.Errorf("err %v, want %v", err, c.want)
			}
		})
	}
}

func TestActionTimeout(t *testing.T) {
	clock := newClock()
	s := sessionWith(clock, nil)
	if got := s.GetActionTimeout(); !got.Equal(startTime().Add(DefaultActionTimeout)) {
		t.Errorf("default timeout %v", got)
	}
	s.SetActionTimeout(5 * time.Second)
	clock.Advance(time.Second)
	if got := s.GetActionTimeout(); !got.Equal(startTime().Add(6 * time.Second)) {
		t.Errorf("custom timeout %v", got)
	}
}

func TestSessionAccessors(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	dev := testkit.NewFakeDevice("d1")
	s := &Session{SessionID: "sid", logger: logrus.NewEntry(logger)}
	if s.GetDevice() != nil || s.GetDeviceLock() != nil {
		t.Error("an unlocked session has no device")
	}
	lock := &manager.DeviceLock{Device: dev}
	s.SetDeviceLock(lock)
	if s.GetDeviceLock() != lock || s.GetDevice() != dev {
		t.Error("lock not returned")
	}
	s.SetLastAccess(startTime())
	if !s.GetLastAccess().Equal(startTime()) || s.GetSessionID() != "sid" || s.GetAppParameter() != nil || s.GetStorage() != nil {
		t.Error("accessors disagree")
	}
	if got := s.GetLogger().Data["session"]; got != "sid" {
		t.Errorf("logger field %v", got)
	}
	s.HandleDisconnect()
}

func TestDeviceDisconnectedLogsOnlyWhileRunning(t *testing.T) {
	logger, hook := logtest.NewNullLogger()
	s := &Session{logger: logrus.NewEntry(logger)}
	s.DeviceDisconnected()
	if len(hook.Entries) != 1 || hook.LastEntry().Message != "Device Crashed!" {
		t.Fatalf("entries %v", hook.Entries)
	}
	_ = s.Close()
	s.DeviceDisconnected()
	if len(hook.Entries) != 1 {
		t.Errorf("a closed session must not log a crash: %v", hook.Entries)
	}
}

func TestCloseStopsRecording(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	fs := newMemFS()
	rec := &Recorder{Storage: NewStorage(deps(fs), "logs", "sid"), Device: dev}
	s := &Session{Recorder: rec}
	if err := rec.Start(); err != nil {
		t.Fatal(err)
	}
	if calls := dev.CallsTo("StartRecording"); len(calls) != 1 || calls[0].Args[0] != sessionDir {
		t.Errorf("start calls %v", calls)
	}
	if err := s.Close(); err != nil || len(dev.CallsTo("StopRecording")) != 1 {
		t.Errorf("close err %v", err)
	}
	dev.FailWith("StopRecording", errDisk)
	if err := s.Close(); !errors.Is(err, errDisk) {
		t.Errorf("close err %v, want %v", err, errDisk)
	}
}
