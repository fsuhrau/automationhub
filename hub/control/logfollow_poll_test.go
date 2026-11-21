package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// ticker replaces the wait between two reads: it records the wait, runs the
// step of the tick, and fires at once. The last step ends the test by
// cancelling the context, so its tick never fires.
type ticker struct {
	waits []time.Duration
	steps []func()
}

func (k *ticker) after(d time.Duration) <-chan time.Time {
	if i := len(k.waits); i < len(k.steps) {
		k.steps[i]()
	}
	k.waits = append(k.waits, d)
	if n := len(k.steps); n > 0 && len(k.waits) >= n {
		return nil
	}
	fired := make(chan time.Time, 1)
	fired <- time.Time{}
	return fired
}

func followService(k *ticker) *Service {
	s := New(nil)
	s.after = k.after
	return s
}

func TestFollowLogsSendsTheBacklogAndThenNewLines(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	dev.LogLines = []string{"11-09 12:00:00.100 I/A( 1): one", "11-09 12:00:00.200 I/A( 1): two"}
	ctx, cancel := context.WithCancel(context.Background())
	k := &ticker{steps: []func(){
		func() { dev.LogLines = append(dev.LogLines, "11-09 12:00:00.300 I/A( 1): three") },
		func() {}, // a read without a new line sends nothing
		cancel,
	}}
	var batches [][]string

	err := followService(k).FollowLogs(ctx, dev, LogFollow{Backlog: 10, OnLines: func(lines []string) error {
		batches = append(batches, lines)
		return nil
	}})

	if err != nil || len(batches) != 2 || len(batches[0]) != 2 || len(batches[1]) != 1 {
		t.Fatalf("batches %v, err %v", batches, err)
	}
	if len(dev.CallsTo("ReadLogs")) != 1 || len(dev.CallsTo("ReadLogsSince")) != 2 {
		t.Fatalf("calls %v", dev.Calls())
	}
}

func TestFollowLogsBacksOffAfterReadErrors(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	dev.FailWith("ReadLogs", errors.New("node gone"))
	stop := errors.New("stop following")
	k := &ticker{}
	var seen int

	err := followService(k).FollowLogs(context.Background(), dev, LogFollow{OnError: func(error) error {
		seen++
		if seen == 3 {
			return stop
		}
		return nil
	}})

	if !errors.Is(err, stop) || seen != 3 {
		t.Fatalf("err %v after %d errors", err, seen)
	}
	if want := []time.Duration{2 * time.Second, 4 * time.Second}; k.waits[0] != want[0] || k.waits[1] != want[1] {
		t.Fatalf("waits %v, want %v", k.waits, want)
	}
}

func TestFollowLogsStopsWhenTheReceiverFails(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	dev.LogLines = []string{"11-09 12:00:00.100 I/A( 1): one"}
	stop := errors.New("receiver gone")

	err := followService(&ticker{}).FollowLogs(context.Background(), dev, LogFollow{Backlog: 1, OnLines: func([]string) error { return stop }})

	if !errors.Is(err, stop) {
		t.Fatalf("err %v", err)
	}
}

func TestFollowLogsNeedsALogReader(t *testing.T) {
	err := New(nil).FollowLogs(context.Background(), &bareDevice{id: "p"}, LogFollow{})

	if !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("err %v", err)
	}
}

// readerOnly reads logs but can't follow them.
type readerOnly struct {
	bareDevice
	lines []string
}

func (r *readerOnly) ReadLogs(int) ([]string, error) { return r.lines, nil }

func TestReadLogsSinceFallsBackToTheRecentLines(t *testing.T) {
	s := New(nil)
	reader := &readerOnly{lines: []string{"recent"}}

	got, err := s.readLogsSince(reader, "11-09 12:00:00.100")
	if err != nil || len(got) != 1 || got[0] != "recent" {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := s.readLogsSince(&bareDevice{id: "p"}, ""); !errors.Is(err, device.ErrUnsupported) {
		t.Fatalf("a device without logs: %v", err)
	}
}
