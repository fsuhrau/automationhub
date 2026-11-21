package basetest

import (
	"bytes"
	"errors"
	"image"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"

	_ "image/png"
)

type handler func(*action.Response)

func (h handler) OnActionResponse(_ interface{}, r *action.Response) { h(r) }

func TestAppManagerAnswersTheActionsOfARun(t *testing.T) {
	m := &AppManager{TestPassed: false, Tests: []action.Test{{Method: "m"}}, Screenshot: PNG()}
	dev := testkit.NewFakeDevice("a")
	var got []*action.Response
	dev.AddActionHandler(handler(func(r *action.Response) { got = append(got, r) }))

	for _, a := range []action.Interface{&action.TestsGet{}, &action.TestStart{}, &action.GetScreenshot{}, &action.GetSceenGraph{}} {
		if err := m.SendAction(dev, a); err != nil {
			t.Fatal(err)
		}
	}
	if len(got) != 3 {
		t.Fatalf("%d answers, want 3 (the scene graph is not answered)", len(got))
	}
	if got[0].Payload.Tests == nil || got[1].ActionType != action.ActionType_ExecutionResult || got[1].Success || got[2].Payload.Screenshot == nil {
		t.Fatalf("answers %+v", got)
	}
}

func TestAppManagerReturnsTheScriptedError(t *testing.T) {
	m := &AppManager{}
	m.FailWith("SendAction", errors.New("down"))
	if err := m.SendAction(testkit.NewFakeDevice("a"), &action.TestsGet{}); err == nil {
		t.Fatal("scripted error lost")
	}
}

func TestPNGIsDecodable(t *testing.T) {
	if _, _, err := image.Decode(bytes.NewReader(PNG())); err != nil {
		t.Fatal(err)
	}
}

func TestNewSetupConnectsTheDevices(t *testing.T) {
	s := NewSetup(t, 2)
	if len(s.Devices) != 2 || len(s.Models) != 2 || s.Deps.ProjectID == "" || s.Test.ID == 0 {
		t.Fatalf("setup %+v", s)
	}
	d := s.Devices[1]
	if d.DeviceState() != device.StateBooted || !d.IsAppConnected() || s.Models[1].Dev != d {
		t.Fatal("device not ready")
	}
}

func TestWatchFinishedReturnsTheEventOfTheRun(t *testing.T) {
	wait := WatchFinished(t)
	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{TestRunID: 1})
	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{TestRunID: 2, Success: true})
	if p := wait(2); !p.Success {
		t.Fatalf("got %+v", p)
	}
}

// the recorder drops events when nobody reads, instead of blocking the run
func TestRecorderDropsEventsWhenFull(t *testing.T) {
	r := make(recorder, 1)
	r.Handle(events.TestRunFinishedPayload{TestRunID: 1})
	r.Handle(events.TestRunFinishedPayload{TestRunID: 2})
	if len(r) != 1 || (<-r).TestRunID != 1 {
		t.Fatal("full recorder did not drop")
	}
}

// failTB turns Fatalf into a panic the test can catch.
type failTB struct {
	testing.TB
	msg string
}

func (f *failTB) Helper() {}
func (f *failTB) Fatalf(format string, args ...any) {
	f.msg = format
	panic(f)
}

func TestWaitFinishedFailsOnTimeout(t *testing.T) {
	tb := &failTB{}
	defer func() {
		if recover() != tb {
			t.Fatal("no failure reported")
		}
	}()
	watcher{got: make(chan events.TestRunFinishedPayload), timeout: time.Millisecond}.wait(tb, 1)
}
