package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

type fixture struct {
	m     *Manager
	clock *testkit.FakeClock
	fs    *memFS
	hook  *logtest.Hook
}

func newFixture() *fixture {
	logger, hook := logtest.NewNullLogger()
	clock, fs := newClock(), newMemFS()
	return &fixture{m: NewManager(logger, Deps{Clock: clock, FS: fs}), clock: clock, fs: fs, hook: hook}
}

func (f *fixture) newSession() manager.Session {
	logger, _ := logtest.NewNullLogger()
	return f.m.CreateNewSession(logrus.NewEntry(logger), nil, nil)
}

func (f *fixture) logged(msg string) bool {
	for _, e := range f.hook.AllEntries() {
		if e.Message == msg {
			return true
		}
	}
	return false
}

func TestCreateNewSessionHasUniqueIDAndStorage(t *testing.T) {
	f := newFixture()
	a, b := f.newSession(), f.newSession()
	if a.GetSessionID() == "" || a.GetSessionID() == b.GetSessionID() {
		t.Errorf("ids %q %q", a.GetSessionID(), b.GetSessionID())
	}
	if len(f.fs.files) != 2 || a.GetStorage() == nil {
		t.Errorf("storage not created: %v", f.fs.files)
	}
	if got, _ := f.m.GetSession(a.GetSessionID()); got != nil {
		t.Error("a created session is not registered yet")
	}
}

func TestAddAndGetSessions(t *testing.T) {
	f := newFixture()
	s := f.newSession()
	f.clock.Advance(time.Hour)
	f.m.AddSession(s)

	if !s.GetLastAccess().Equal(startTime().Add(time.Hour)) {
		t.Errorf("last access %v", s.GetLastAccess())
	}
	if got, err := f.m.GetSession(s.GetSessionID()); err != nil || got != s {
		t.Errorf("get: %v %v", got, err)
	}
	if f.m.GetSessionDetails(s.GetSessionID()) != s || len(f.m.GetSessions()) != 1 {
		t.Error("details or list wrong")
	}
	if _, err := f.m.GetSession("nope"); !errors.Is(err, ErrUnknown) {
		t.Errorf("err %v, want %v", err, ErrUnknown)
	}
	if f.m.GetSessionDetails("nope") != nil {
		t.Error("unknown details must be nil")
	}
}

func TestStopSessionClosesAndForgets(t *testing.T) {
	f := newFixture()
	s := f.newSession()
	f.m.AddSession(s)
	if err := f.m.StopSession(s); err != nil {
		t.Fatal(err)
	}
	if f.fs.closeCalls != 1 || len(f.m.GetSessions()) != 0 {
		t.Errorf("close calls %d, sessions %d", f.fs.closeCalls, len(f.m.GetSessions()))
	}
	if err := f.m.StopSession(nil); err != nil {
		t.Errorf("stopping nil: %v", err)
	}
}

func TestStopSessionSurvivesCloseErrorAndMissingStorage(t *testing.T) {
	f := newFixture()
	dev := testkit.NewFakeDevice("d1")
	dev.FailWith("StopRecording", errDisk)
	s := &Session{SessionID: "x", Recorder: &Recorder{Device: dev}}
	f.m.AddSession(s)
	if err := f.m.StopSession(s); err != nil {
		t.Fatal(err)
	}
	if f.m.GetSessionDetails("x") != nil {
		t.Error("session must be forgotten")
	}
	found := false
	for _, e := range f.hook.AllEntries() {
		found = found || e.Message == "Session can't be cleaned: disk full"
	}
	if !found {
		t.Errorf("close error not logged: %v", f.hook.AllEntries())
	}
}

func TestRunStopsExpiredSessions(t *testing.T) {
	f := newFixture()
	idle, busy := f.newSession(), f.newSession()
	f.m.AddSession(idle)
	f.m.AddSession(busy)

	ctx, cancel := context.WithCancel(context.Background())
	f.m.Run(ctx)
	eventually(t, "first cleanup", func() bool { return f.clock.Pending() == 1 })
	if len(f.m.GetSessions()) != 2 {
		t.Fatal("fresh sessions must stay")
	}

	busy.SetLastAccess(startTime().Add(Timeout))
	f.clock.Advance(Timeout + time.Second)
	eventually(t, "idle session stopped", func() bool { return f.m.GetSessionDetails(idle.GetSessionID()) == nil })
	if f.m.GetSessionDetails(busy.GetSessionID()) == nil {
		t.Error("recently used session must stay")
	}
	if !f.logged("session " + idle.GetSessionID() + " expired") {
		t.Error("expiry not logged")
	}

	eventually(t, "next wait", func() bool { return f.clock.Pending() == 1 })
	cancel()
	eventually(t, "loop stopped", func() bool { return f.logged("Stop SessionManager") })
}

func TestRunWithCanceledContextStopsAtOnce(t *testing.T) {
	f := newFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f.m.Run(ctx)
	eventually(t, "loop stopped", func() bool { return f.logged("Stop SessionManager") })
}
