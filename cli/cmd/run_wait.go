package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fsuhrau/automationhub/events"
	sse "github.com/r3labs/sse/v2"
)

// eventStream subscribes channels to the named server-sent event streams.
type eventStream interface {
	Subscribe(ctx context.Context, stream string, ch chan *sse.Event) error
}

// sseStream subscribes with one client per stream: the client does not
// support concurrent subscriptions.
type sseStream struct {
	apiURL string
	delay  time.Duration
}

const (
	reconnectAttempts = 3
	reconnectDelay    = time.Second
)

func newSSEStream(apiURL string) eventStream {
	return sseStream{apiURL: apiURL, delay: reconnectDelay}
}

func (s sseStream) Subscribe(ctx context.Context, stream string, ch chan *sse.Event) error {
	client := sse.NewClient(s.apiURL + "/api/sse")
	client.ReconnectStrategy = &limitedRetries{left: reconnectAttempts, delay: s.delay}
	return client.SubscribeChanWithContext(ctx, stream, ch)
}

// limitedRetries is a backoff strategy that gives up after a few attempts;
// the client's default retries for 15 minutes.
type limitedRetries struct {
	left  int
	delay time.Duration
}

// NextBackOff is the delay before the next attempt, -1 (stop) when none is left.
func (r *limitedRetries) NextBackOff() time.Duration {
	if r.left <= 0 {
		return -1
	}
	r.left--
	return r.delay
}

func (r *limitedRetries) Reset() {}

// waiter prints the log of a run to out while it waits for the result.
type waiter struct {
	stream eventStream
	out    io.Writer
}

// await returns whether the run succeeded once it finished; ctx ends the wait.
func (w waiter) await(ctx context.Context, runID uint) (bool, error) {
	finished := make(chan *sse.Event)
	logs := make(chan *sse.Event)
	if err := w.stream.Subscribe(ctx, fmt.Sprintf("test_run_%d_finished", runID), finished); err != nil {
		return false, fmt.Errorf("subscribe to the result: %w", err)
	}
	if err := w.stream.Subscribe(ctx, fmt.Sprintf("test_run_%d_log", runID), logs); err != nil {
		return false, fmt.Errorf("subscribe to the log: %w", err)
	}
	for {
		select {
		case event := <-finished:
			return decodeFinished(event)
		case event := <-logs:
			printLogEntry(w.out, event)
		case <-ctx.Done():
			return false, fmt.Errorf("waiting for test run %d: %w", runID, ctx.Err())
		}
	}
}

func decodeFinished(event *sse.Event) (bool, error) {
	var payload events.TestRunFinishedPayload
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return false, fmt.Errorf("decode result: %w", err)
	}
	return payload.Success, nil
}

func printLogEntry(out io.Writer, event *sse.Event) {
	var entry events.NewTestLogEntryPayload
	if err := json.Unmarshal(event.Data, &entry); err != nil {
		fprintf(out, "log: %s\n", event.Data)
		return
	}
	fprintf(out, "log: %v\n", entry)
}
