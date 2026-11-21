package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func newTestWizard(input string) (*wizard, *bytes.Buffer) {
	var out bytes.Buffer
	return newWizard(strings.NewReader(input), &out), &out
}

func TestChooseAsksUntilTheAnswerIsValid(t *testing.T) {
	w, out := newTestWizard("maybe\n\nn\n")

	if got := w.choose("Q", "y", "n"); got != "n" {
		t.Errorf("got %q", got)
	}
	if strings.Count(out.String(), "Q -> ") != 3 {
		t.Errorf("prompts: %q", out)
	}
}

func TestAskYes(t *testing.T) {
	tests := map[string]bool{"y\n": true, "n\n": false}
	for input, want := range tests {
		w, out := newTestWizard(input)
		if got := w.askYes("Go?"); got != want {
			t.Errorf("%q: got %v", input, got)
		}
		if !strings.Contains(out.String(), "Go? (y/n) -> ") {
			t.Errorf("prompt: %q", out)
		}
	}
}

func TestReadIPAsksUntilTheIPIsValid(t *testing.T) {
	w, _ := newTestWizard("999.1.1.1\nabc\n10.0.0.2\n")

	if got := w.readIP("ip"); got != "10.0.0.2" {
		t.Errorf("got %q", got)
	}
}

func TestEndOfInputStopsEveryQuestion(t *testing.T) {
	w, _ := newTestWizard("")

	if w.choose("Q", "y") != "" || w.readIP("ip") != "" || w.readLine("x") != "" {
		t.Error("answered without input")
	}
	if w.err == nil || !strings.Contains(w.err.Error(), "read input") {
		t.Errorf("err = %v", w.err)
	}
}

func TestAnswerWithoutNewlineAtTheEndIsAccepted(t *testing.T) {
	w, _ := newTestWizard("y")

	if !w.askYes("Q") || w.err != nil {
		t.Errorf("err = %v", w.err)
	}
}

func TestWriteErrorStopsTheWizard(t *testing.T) {
	w := newWizard(strings.NewReader("y\n"), failingWriter{})

	w.say("a")
	w.say("b")

	if w.err == nil || !strings.Contains(w.err.Error(), "write output") {
		t.Errorf("err = %v", w.err)
	}
	if _, err := w.collect(); err == nil {
		t.Error("collect succeeded")
	}
}

func collectFrom(t *testing.T, input string) config.Service {
	t.Helper()
	w, _ := newTestWizard(input)
	cfg, err := w.collect()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestCollectEnablesEveryManagerOnYes(t *testing.T) {
	cfg := collectFrom(t, "y\ny\ny\ny\ny\ny\nn\ny\n")

	if !cfg.Autodetect || cfg.HostIP != "" || len(cfg.DeviceManager) != 4 {
		t.Fatalf("cfg = %+v", cfg)
	}
	for name, m := range cfg.DeviceManager {
		wantShot := name == androiddevice.Manager || name == iosdevice.Manager
		if !m.Enabled || m.UseOSScreenshot != wantShot {
			t.Errorf("%s = %+v", name, m)
		}
	}
}

func TestCollectTakesAStaticIPAndNoManagers(t *testing.T) {
	cfg := collectFrom(t, "n\n10.0.0.9\nn\nn\nn\nn\n")

	if cfg.Autodetect || cfg.HostIP != "10.0.0.9" || len(cfg.DeviceManager) != 0 {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestIOSManagerGetsTheIntegrationBundle(t *testing.T) {
	w, _ := newTestWizard("y\nn\n")
	cfg := config.Service{DeviceManager: map[string]config.Manager{}}

	w.askManager(&cfg, managerPrompts()[1])

	if m := cfg.DeviceManager[iosdevice.Manager]; m.WebDriver == nil || m.WebDriver.BundleID != iosBundleID {
		t.Errorf("manager = %+v", m)
	}
}
