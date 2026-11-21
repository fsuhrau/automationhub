package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAwaitRun(t *testing.T) {
	finished := func(data string) map[string][]*sseEvent {
		return map[string][]*sseEvent{"test_run_5_finished": {{Data: []byte(data)}}}
	}
	boom := errors.New("refused")
	tests := map[string]struct {
		stream      fakeStream
		wantSuccess bool
		wantErr     string
	}{
		"success":              {stream: fakeStream{events: finished(`{"success":true}`)}, wantSuccess: true},
		"failure":              {stream: fakeStream{events: finished(`{"success":false}`)}},
		"result is no json":    {stream: fakeStream{events: finished(`nope`)}, wantErr: "decode result"},
		"result subscribe":     {stream: fakeStream{fail: map[string]error{"test_run_5_finished": boom}}, wantErr: "subscribe to the result"},
		"log subscribe":        {stream: fakeStream{fail: map[string]error{"test_run_5_log": boom}}, wantErr: "subscribe to the log"},
		"nothing ever arrives": {stream: fakeStream{}, wantErr: "waiting for test run 5"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			success, err := waiter{tc.stream, &bytes.Buffer{}}.await(ctx, 5)

			if success != tc.wantSuccess || (tc.wantErr == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), tc.wantErr)) {
				t.Errorf("success = %v, err = %v", success, err)
			}
		})
	}
}

func TestAwaitRunPrintsTheLogEntries(t *testing.T) {
	stream := fakeStream{events: map[string][]*sseEvent{
		"test_run_5_log":      {{Data: []byte(`{"testRunId":5,"entry":"step"}`)}, {Data: []byte(`not json`)}},
		"test_run_5_finished": {},
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var out bytes.Buffer

	_, err := waiter{stream, &out}.await(ctx, 5)

	if err == nil {
		t.Error("expected the wait to time out")
	}
	for _, want := range []string{"log: {5 step}", "log: not json"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("out lacks %q:\n%s", want, &out)
		}
	}
}

func TestLimitedRetriesGiveUpAfterTheLastAttempt(t *testing.T) {
	r := &limitedRetries{left: 2, delay: time.Second}

	got := []time.Duration{r.NextBackOff(), r.NextBackOff(), r.NextBackOff()}
	r.Reset()

	if got[0] != time.Second || got[1] != time.Second || got[2] != -1 {
		t.Errorf("delays = %v", got)
	}
}

func TestSSEStreamFailsAfterTheRetriesWhenTheHubHasNoStream(t *testing.T) {
	hub := newFakeHub(t)
	stream := sseStream{apiURL: hub.URL, delay: time.Millisecond}

	err := stream.Subscribe(context.Background(), "unknown", make(chan *sseEvent))

	if err == nil {
		t.Error("subscribed to a stream that does not exist")
	}
}
