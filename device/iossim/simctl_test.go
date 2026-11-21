package iossim

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestParseSimulators(t *testing.T) {
	sims, err := parseSimulators([]byte(listJSON))
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(sims, func(i, j int) bool { return sims[i].UDID < sims[j].UDID })
	want := []simInfo{
		{UDID: "AAAA-1111", Name: "iPhone 15", State: device.StateBooted, OSName: "iphonesimulator", OSVersion: "17.0"},
		{UDID: "BBBB-2222", Name: "iPhone 15 Pro", State: device.StateShutdown, OSName: "iphonesimulator", OSVersion: "17.0"},
		{UDID: "CCCC-3333", Name: "Apple TV", State: device.StateShutdown, OSName: "tvOS", OSVersion: "17.0"},
	}
	if !reflect.DeepEqual(sims, want) {
		t.Fatalf("got %+v, want %+v", sims, want)
	}
}

func TestParseSimulatorsInvalidJSON(t *testing.T) {
	if _, err := parseSimulators([]byte("{")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestSimctlList(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		err     error
		wantLen int
		wantErr bool
	}{
		{name: "ok", out: listJSON, wantLen: 3},
		{name: "run error", err: errBoom, wantErr: true},
		{name: "bad json", out: "x", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testkit.FakeRunner{}
			r.Script(listLine, []byte(tt.out), tt.err)
			sims, err := simctl{runner: r}.list()
			if (err != nil) != tt.wantErr || len(sims) != tt.wantLen {
				t.Fatalf("got %d sims, err %v", len(sims), err)
			}
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("error not wrapped: %v", err)
			}
			if got := r.Commands(); len(got) != 1 || got[0] != listLine {
				t.Fatalf("commands: %v", got)
			}
		})
	}
}

func TestParseSimState(t *testing.T) {
	tests := map[string]device.State{"Booted": device.StateBooted, "Shutdown": device.StateShutdown, "Creating": device.StateUnknown}
	for in, want := range tests {
		if got := parseSimState(in); got != want {
			t.Errorf("%s: got %v, want %v", in, got, want)
		}
	}
}
