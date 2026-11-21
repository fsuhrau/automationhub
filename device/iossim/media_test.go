package iossim

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	shotLine = "xcrun simctl io AAAA-1111 screenshot --type=png -"
	recLine  = "xcrun simctl io AAAA-1111 recordVideo ./run1.mp4"
)

func TestGetScreenshot(t *testing.T) {
	good := pngBytes(t, 4, 7)
	tests := []struct {
		name    string
		out     []byte
		err     error
		wantErr bool
	}{
		{name: "png", out: good},
		{name: "run error", err: errBoom, wantErr: true},
		{name: "not an image", out: []byte("nope"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testkit.FakeRunner{}
			r.Script(shotLine, tt.out, tt.err)
			data, w, h, err := newTestDevice(r, false).GetScreenshot()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if !tt.wantErr && (w != 4 || h != 7 || len(data) != len(good)) {
				t.Fatalf("got %dx%d, %d bytes", w, h, len(data))
			}
			if got := r.Commands(); len(got) != 1 || got[0] != shotLine {
				t.Fatalf("commands: %v", got)
			}
		})
	}
}

func TestRecording(t *testing.T) {
	r := &testkit.FakeRunner{}
	r.Script(recLine, nil, nil)
	d := newTestDevice(r, false)
	if err := d.StopRecording(); err != nil {
		t.Fatalf("stop without recording: %v", err)
	}
	if err := d.StartRecording("run1"); err != nil {
		t.Fatal(err)
	}
	if err := d.StopRecording(); err != nil {
		t.Fatal(err)
	}
	if got := r.Commands(); len(got) != 1 || got[0] != recLine {
		t.Fatalf("commands: %v", got)
	}
}

func TestStartRecordingFailure(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}, false) // unscripted: Start fails
	if err := d.StartRecording("run1"); err == nil {
		t.Fatal("expected an error")
	}
}

type killFails struct{}

func (killFails) Wait() error { return nil }
func (killFails) Kill() error { return errBoom }

func TestStopRecordingKillError(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}, false)
	d.recording = killFails{}
	if err := d.StopRecording(); !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
}
