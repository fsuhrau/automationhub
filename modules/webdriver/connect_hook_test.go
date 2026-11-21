package webdriver

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func neverAfter(time.Duration) <-chan time.Time { return nil }

func firedAfter(time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- time.Time{}
	return ch
}

func TestParseServerURL(t *testing.T) {
	for line, want := range map[string]string{
		"x ServerURLHere->http://10.0.0.2:8100<-ServerURLHere": "http://10.0.0.2:8100",
	} {
		if got, ok := ParseServerURL(line); !ok || got != want {
			t.Errorf("ParseServerURL(%q) = %q, %v", line, got, ok)
		}
	}
	for _, line := range []string{"nothing", "ServerURLHere-><-ServerURLHere", "ServerURLHere only"} {
		if _, ok := ParseServerURL(line); ok {
			t.Errorf("ParseServerURL(%q) parsed", line)
		}
	}
}

func TestAwaitReturnsTheLoggedURL(t *testing.T) {
	hook := &WDAConnectHook{}
	got, err := hook.Await(time.Minute, neverAfter, func(func(error)) {
		_ = hook.Fire(&logrus.Entry{Message: "ServerURLHere->http://h:1<-ServerURLHere"})
		_ = hook.Fire(&logrus.Entry{Message: "ServerURLHere->http://second:2<-ServerURLHere"}) // buffer full, dropped
	})
	if err != nil || got != "http://h:1" {
		t.Fatalf("Await = %q, %v", got, err)
	}
}

func TestAwaitFailsWhenTheRunnerStops(t *testing.T) {
	hook := &WDAConnectHook{}
	cause := errors.New("boom")
	_, err := hook.Await(time.Minute, neverAfter, func(fail func(error)) {
		fail(cause)
		fail(errors.New("second failure is dropped"))
	})
	if !errors.Is(err, cause) {
		t.Fatalf("err = %v", err)
	}
}

func TestAwaitTimesOut(t *testing.T) {
	hook := &WDAConnectHook{}
	if _, err := hook.Await(time.Second, firedAfter, func(func(error)) {}); err == nil {
		t.Fatal("expected a timeout")
	}
}

func TestFireIgnoresOtherLinesAndNoWaiter(t *testing.T) {
	hook := &WDAConnectHook{}
	if err := hook.Fire(&logrus.Entry{Message: "unrelated"}); err != nil {
		t.Fatal(err)
	}
	if err := hook.Fire(&logrus.Entry{Message: "ServerURLHere->http://h:1<-ServerURLHere"}); err != nil {
		t.Fatal(err)
	}
	if len(hook.Levels()) != len(logrus.AllLevels) {
		t.Fatal("the hook must see all levels")
	}
}

func TestHookWorksAsALogrusHook(t *testing.T) {
	hook := &WDAConnectHook{}
	logger := logrus.New()
	logger.AddHook(hook)
	logger.SetOutput(discard{})
	got, err := hook.Await(time.Minute, neverAfter, func(func(error)) {
		logger.Info("ServerURLHere->http://h:3<-ServerURLHere")
	})
	if err != nil || got != "http://h:3" {
		t.Fatalf("Await = %q, %v", got, err)
	}
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
