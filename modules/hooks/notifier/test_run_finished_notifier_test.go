package notifier

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/storage/models"
)

type fakeHook struct{ out chan hooks.Message }

func (f fakeHook) Send(msg hooks.Message) { f.out <- msg }

func newTestRun(testID, id uint, test *models.Test) *models.TestRun {
	run := &models.TestRun{TestID: testID, Test: test}
	run.ID = id
	return run
}

func TestBuildMessage(t *testing.T) {
	got := buildMessage(events.TestRunFinishedPayload{Succeeded: 3, Unstable: 2, Failed: 1})
	want := "Results:\n- Successful: 3\n- Unstable: 2\n- Failed: 1 "
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildTitle(t *testing.T) {
	tests := map[string]struct {
		payload events.TestRunFinishedPayload
		want    string
	}{
		"test name":        {events.TestRunFinishedPayload{TestName: "smoke"}, "TestRun finished: smoke"},
		"name of the run":  {events.TestRunFinishedPayload{TestRun: newTestRun(1, 1, &models.Test{Name: "login"})}, "TestRun finished: login"},
		"run without test": {events.TestRunFinishedPayload{TestRun: newTestRun(1, 1, nil)}, "TestRun finished"},
		"no run":           {events.TestRunFinishedPayload{}, "TestRun finished"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := buildTitle(tc.payload); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetURLUsesTheInjectedHostIP(t *testing.T) {
	payload := events.TestRunFinishedPayload{ProjectID: "p", AppID: 4, TestRun: newTestRun(7, 9, nil)}
	if got, want := getUrl(payload, "10.0.0.5"), "http://10.0.0.5:8002/project/p/app/4/test/7/run/9"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := getUrl(events.TestRunFinishedPayload{}, "10.0.0.5"); got != "" {
		t.Errorf("without a run got %q", got)
	}
}

func TestGetLevel(t *testing.T) {
	tests := map[string]struct {
		payload events.TestRunFinishedPayload
		want    hooks.Level
	}{
		"unstable wins": {events.TestRunFinishedPayload{Unstable: 1, Failed: 1, Succeeded: 1}, hooks.LevelUnstable},
		"failed":        {events.TestRunFinishedPayload{Failed: 1, Succeeded: 1}, hooks.LevelError},
		"succeeded":     {events.TestRunFinishedPayload{Succeeded: 1}, hooks.LevelSuccess},
		"nothing":       {events.TestRunFinishedPayload{}, hooks.LevelUnstable},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getLevel(tc.payload); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRegisteredListenerSendsToEveryHookWhenARunFinishes(t *testing.T) {
	out := make(chan hooks.Message, 2)
	RegisterEventTestRunFinishedListener([]hooks.Hook{fakeHook{out}, fakeHook{out}}, "1.2.3.4")

	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{
		ProjectID: "p", AppID: 1, TestName: "smoke", Failed: 1, TestRun: newTestRun(2, 3, nil),
	})

	for i := 0; i < 2; i++ {
		select {
		case got := <-out:
			want := hooks.Message{
				Title: "TestRun finished: smoke",
				Text:  buildMessage(events.TestRunFinishedPayload{Failed: 1}),
				Link:  "http://1.2.3.4:8002/project/p/app/1/test/2/run/3",
				Level: hooks.LevelError,
			}
			if got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("hook was not called")
		}
	}
}
