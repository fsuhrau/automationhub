package unityeditor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

// editorComesUp creates the instance file and boots the editor at the given
// offsets from the start.
func editorComesUp(dev *Device, dir string, fileAfter, bootAfter time.Duration) func(time.Time) {
	return func(now time.Time) {
		if now.Sub(epochTime()) >= fileAfter {
			_ = os.WriteFile(filepath.Join(dir, instanceFileName), nil, 0o644)
		}
		if now.Sub(epochTime()) >= bootAfter {
			dev.setState(device.StateBooted)
		}
	}
}

func TestStartDeviceLaunchesTheEditor(t *testing.T) {
	dir := project(t)
	var dev *Device
	clock := newTickClock(t, func(now time.Time) { editorComesUp(dev, dir, 10*time.Second, 15*time.Second)(now) })
	env := newEnv(clock)
	dev = env.addEditor("e1", dir)
	line := launchLine(dir, epochTime())
	env.runner.Script(line, nil, nil)

	if err := env.handler.StartDevice("e1"); err != nil {
		t.Fatalf("StartDevice: %v", err)
	}
	if got := env.runner.Commands(); len(got) != 1 || got[0] != line {
		t.Fatalf("commands = %q, want [%q]", got, line)
	}
	if dev.process == nil || dev.instanceLogFile == "" {
		t.Fatal("the editor process and log file are kept for StopDevice")
	}
}

func TestStartDeviceRemovesAStaleInstanceFile(t *testing.T) {
	dir := project(t)
	stale := filepath.Join(dir, instanceFileName)
	if err := os.WriteFile(stale, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	env := newEnv(newTickClock(t, nil))
	env.addEditor("e1", dir)
	env.runner.Script(launchLine(dir, epochTime()), nil, nil)

	err := env.handler.StartDevice("e1")
	if !errors.Is(err, errStartTimeout) {
		t.Fatalf("err = %v, want errStartTimeout: the old %s must not count", err, instanceFileName)
	}
}

func TestStartDeviceTimeouts(t *testing.T) {
	tests := []struct {
		name      string
		fileAfter time.Duration
		bootAfter time.Duration
	}{
		{"instance file never appears", time.Hour, time.Hour},
		{"editor never boots", time.Second, time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := project(t)
			var dev *Device
			env := newEnv(newTickClock(t, func(now time.Time) { editorComesUp(dev, dir, tt.fileAfter, tt.bootAfter)(now) }))
			dev = env.addEditor("e1", dir)
			env.runner.Script(launchLine(dir, epochTime()), nil, nil)

			if err := env.handler.StartDevice("e1"); !errors.Is(err, errStartTimeout) {
				t.Fatalf("err = %v, want errStartTimeout", err)
			}
			if got := env.handler.clock.Now().Sub(epochTime()); got <= StartupTimeout {
				t.Fatalf("gave up after %v, want more than %v", got, StartupTimeout)
			}
		})
	}
}

func TestStartDeviceErrors(t *testing.T) {
	t.Run("unknown device", func(t *testing.T) {
		if err := newEnv(nil).handler.StartDevice("x"); !errors.Is(err, device.DeviceNotFoundError) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("device without config has no project", func(t *testing.T) {
		env := newEnv(nil)
		env.handler.Add(newDevice(infoOf("e1"), env.clock))
		if err := env.handler.StartDevice("e1"); !errors.Is(err, ErrorGetProjectVersion) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("project without version file", func(t *testing.T) {
		env := newEnv(nil)
		env.addEditor("e1", t.TempDir())
		if err := env.handler.StartDevice("e1"); !errors.Is(err, ErrorGetProjectVersion) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("editor does not start", func(t *testing.T) {
		env := newEnv(nil)
		env.addEditor("e1", project(t))
		if err := env.handler.StartDevice("e1"); err == nil {
			t.Fatal("an unscripted launch must fail")
		}
	})
}

func TestStopDevice(t *testing.T) {
	t.Run("unknown device is fine", func(t *testing.T) {
		if err := newEnv(nil).handler.StopDevice("x"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("kills the editor once", func(t *testing.T) {
		env := newEnv(nil)
		dev := env.addEditor("e1", "")
		process := &testkit.FakeProcess{}
		dev.setLaunch(process, "log")
		for i := 0; i < 2; i++ {
			if err := env.handler.StopDevice("e1"); err != nil {
				t.Fatal(err)
			}
		}
		if process.Killed() != 1 || dev.process != nil || dev.logFile() != "log" {
			t.Fatalf("killed = %d, process = %v", process.Killed(), dev.process)
		}
	})
	t.Run("kill failure is returned", func(t *testing.T) {
		env := newEnv(nil)
		dev := env.addEditor("e1", "")
		dev.setLaunch(failingProcess{}, "")
		if err := env.handler.StopDevice("e1"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
		if dev.process == nil {
			t.Fatal("the process stays so StopDevice can be retried")
		}
	})
}

func TestEditorBinary(t *testing.T) {
	if got := editorBinary("darwin", "/U", "1.2.3"); got != "/U/1.2.3/Unity.app/Contents/MacOS/Unity" {
		t.Errorf("darwin = %q", got)
	}
	if got := editorBinary("linux", "/U", "1.2.3"); got != "" {
		t.Errorf("linux = %q, want empty", got)
	}
	_ = GetUnityEditorPath("/U", "1.2.3")
}

func TestGetUsedUnityVersion(t *testing.T) {
	if got, err := GetUsedUnityVersion(project(t)); err != nil || got != testVersion {
		t.Fatalf("got %q, %v", got, err)
	}
	broken := project(t)
	file := filepath.Join(broken, "ProjectSettings", "ProjectVersion.txt")
	if err := os.WriteFile(file, []byte("nothing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GetUsedUnityVersion(broken); !errors.Is(err, ErrorGetProjectVersion) {
		t.Fatalf("err = %v", err)
	}
}
