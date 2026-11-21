package hostinfo

import (
	"context"
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

const ioregLine = "ioreg -rd1 -c IOPlatformExpertDevice"

const (
	hardwareProfile = `Hardware:

    Hardware Overview:

      Model Name: MacBook Pro
      Model Identifier: MacBookPro18,1
      Serial Number (system): C02XYZ
`
	displaysProfile = `Graphics/Displays:

    Apple M1 Pro:

      Chipset Model: Apple M1 Pro
      Metal Support: Metal 3
`
	ioregOut = `+-o Root  <class IORegistryEntry>
    "IOPlatformUUID" = "54DECB62-3993-4031-9C6A-18CE048CC63C"
`
)

func darwinHost(r *testkit.FakeRunner) *Host { return New(Env{Runner: r, GOOS: "darwin"}) }

func script(r *testkit.FakeRunner, line, out string) { r.Script(line, []byte(out), nil) }

func TestEnvDefaults(t *testing.T) {
	e := Env{}.WithDefaults()
	if e.Runner == nil || e.Clock == nil || e.GOOS == "" {
		t.Fatalf("defaults not applied: %+v", e)
	}
	fake := &testkit.FakeRunner{}
	if got := (Env{Runner: fake, GOOS: "linux"}).WithDefaults(); got.Runner != device.CommandRunner(fake) || got.GOOS != "linux" {
		t.Fatalf("set fields must stay: %+v", got)
	}
}

func TestSystemValues(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name    string
		goos    string
		scripts map[string]string
		runErr  error // returned by every scripted command
		get     func(*Host, context.Context) (string, error)
		want    string
		wantErr error
		cmds    []string
	}{
		{name: "os name darwin", goos: "darwin", scripts: map[string]string{"sw_vers -productName": "macOS\n"}, get: (*Host).OSName, want: "macOS", cmds: []string{"sw_vers -productName"}},
		{name: "os name linux is goos", goos: "linux", get: (*Host).OSName, want: "linux"},
		{name: "os version darwin", goos: "darwin", scripts: map[string]string{"sw_vers -productVersion": "14.5\n"}, get: (*Host).OSVersion, want: "14.5", cmds: []string{"sw_vers -productVersion"}},
		{name: "os version linux unsupported", goos: "linux", get: (*Host).OSVersion, wantErr: device.ErrUnsupported},
		{name: "host name darwin", goos: "darwin", scripts: map[string]string{"sysctl -n kern.hostname": "mac-1\n"}, get: (*Host).HostName, want: "mac-1", cmds: []string{"sysctl -n kern.hostname"}},
		{name: "host name linux", goos: "linux", scripts: map[string]string{"hostname": "box\n"}, get: (*Host).HostName, want: "box", cmds: []string{"hostname"}},
		{name: "uuid", goos: "darwin", scripts: map[string]string{ioregLine: ioregOut}, get: (*Host).HardwareUUID, want: "54DECB62-3993-4031-9C6A-18CE048CC63C", cmds: []string{ioregLine}},
		{name: "uuid missing", goos: "darwin", scripts: map[string]string{ioregLine: "nothing\n"}, get: (*Host).HardwareUUID, wantErr: ErrNotFound},
		{name: "uuid linux unsupported", goos: "linux", get: (*Host).HardwareUUID, wantErr: device.ErrUnsupported},
		{name: "runner error wrapped", goos: "linux", scripts: map[string]string{"hostname": ""}, runErr: boom, get: (*Host).HostName, wantErr: boom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &testkit.FakeRunner{}
			for line, out := range tc.scripts {
				r.Script(line, []byte(out), tc.runErr)
			}
			got, err := tc.get(New(Env{Runner: r, GOOS: tc.goos}), context.Background())
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if tc.cmds != nil && !equal(r.Commands(), tc.cmds) {
				t.Fatalf("commands %v, want %v", r.Commands(), tc.cmds)
			}
		})
	}
}

func TestOutputIsCached(t *testing.T) {
	r := &testkit.FakeRunner{}
	script(r, "sw_vers -productVersion", "14.5\n")
	h := darwinHost(r)
	for i := 0; i < 3; i++ {
		if _, err := h.OSVersion(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(r.Commands()); n != 1 {
		t.Fatalf("command ran %d times, want 1", n)
	}
}

func TestProfileValues(t *testing.T) {
	r := &testkit.FakeRunner{}
	script(r, "system_profiler SPHardwareDataType", hardwareProfile)
	script(r, "system_profiler SPDisplaysDataType", displaysProfile)
	script(r, "sysctl -n machdep.cpu.brand_string", "Apple M1 Pro\n")
	script(r, "sysctl -n hw.memsize", "17179869184\n")
	h := darwinHost(r)
	ctx := context.Background()
	tests := []struct {
		name string
		get  func(context.Context) (string, error)
		want string
	}{
		{"model", h.ModelIdentifier, "MacBookPro18,1"},
		{"serial", h.SerialNumber, "C02XYZ"},
		{"gpu", h.GPU, "Apple M1 Pro"},
		{"drivers", h.GraphicsDrivers, "Metal 3"},
		{"cpu", h.CPU, "Apple M1 Pro"},
		{"ram", h.RAM, "16 GB"},
	}
	for _, tc := range tests {
		got, err := tc.get(ctx)
		if err != nil || got != tc.want {
			t.Errorf("%s: got %q, %v, want %q", tc.name, got, err, tc.want)
		}
	}
}

func TestProfileErrors(t *testing.T) {
	r := &testkit.FakeRunner{}
	script(r, "system_profiler SPHardwareDataType", "Hardware:\n")
	script(r, "sysctl -n hw.memsize", "lots\n")
	h := darwinHost(r)
	ctx := context.Background()
	if _, err := h.SerialNumber(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("serial: %v, want ErrNotFound", err)
	}
	if _, err := h.RAM(ctx); err == nil {
		t.Error("ram: expected a parse error")
	}
	if _, err := h.GPU(ctx); err == nil {
		t.Error("gpu: expected the unscripted command to fail")
	}
	if _, err := New(Env{Runner: r, GOOS: "linux"}).RAM(ctx); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("ram on linux: %v, want ErrUnsupported", err)
	}
	if _, err := darwinHost(&testkit.FakeRunner{}).RAM(ctx); err == nil {
		t.Error("ram: expected the runner error")
	}
}

func TestParameters(t *testing.T) {
	r := &testkit.FakeRunner{}
	script(r, "system_profiler SPHardwareDataType", hardwareProfile)
	got := darwinHost(r).Parameters(context.Background(), ParamModel, ParamSerial, ParamCPU, "bogus")
	want := map[string]string{ParamModel: "MacBookPro18,1", ParamSerial: "C02XYZ"}
	if len(got) != len(want) || got[ParamModel] != want[ParamModel] || got[ParamSerial] != want[ParamSerial] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
