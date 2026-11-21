package unityeditor

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestNewDeviceIdentity(t *testing.T) {
	ip := net.ParseIP("10.0.0.1")
	d := newDevice(generic.Info{ID: "e1", Name: "n", OSName: "os", OSVersion: "1", Target: "2022"}, nil)
	d.ip = ip
	checks := map[string]bool{
		"id":       d.DeviceID() == "e1",
		"name":     d.DeviceName() == "n",
		"os":       d.DeviceOSName() == "os" && d.DeviceOSVersion() == "1",
		"target":   d.TargetVersion() == "2022",
		"type":     d.DeviceType() == int(models.DeviceTypeUnityEditor),
		"platform": d.PlatformType() == int(models.PlatformTypeEditor),
		"ip":       d.DeviceIP().Equal(ip),
		"timeout":  d.ConnectionTimeout() == ConnectionTimeout,
	}
	for name, ok := range checks {
		if !ok {
			t.Errorf("%s is wrong", name)
		}
	}
}

func TestUnsupportedRoles(t *testing.T) {
	d := &Device{}
	tests := map[string]error{
		"install":         d.InstallApp(nil),
		"uninstall":       d.UninstallApp(nil),
		"start recording": d.StartRecording("p"),
		"stop recording":  d.StopRecording(),
	}
	for name, err := range tests {
		if !errors.Is(err, device.ErrUnsupported) {
			t.Errorf("%s: err = %v, want device.ErrUnsupported", name, err)
		}
	}
}

func TestNoOpRoles(t *testing.T) {
	d := &Device{}
	installed, err := d.IsAppInstalled(nil)
	if !installed || err != nil {
		t.Fatalf("IsAppInstalled = %v, %v", installed, err)
	}
	if d.HasFeature("x") || d.IsAppConnected() {
		t.Fatal("a new editor has no feature and no app connection")
	}
	d.Execute("x")
	d.RunNativeScript(nil)
}

func TestSendCommand(t *testing.T) {
	t.Run("start and stop are queued as json", func(t *testing.T) {
		d := &Device{sendCh: make(chan []byte, 2)}
		if err := d.StartApp(nil, nil, "s1", "node:1"); err != nil {
			t.Fatal(err)
		}
		if err := d.StopApp(nil); err != nil {
			t.Fatal(err)
		}
		if got := string(<-d.sendCh); got != `{"Action":"start","HostUrl":"node:1","SessionID":"s1"}` {
			t.Errorf("start = %s", got)
		}
		if got := string(<-d.sendCh); got != `{"Action":"stop"}` {
			t.Errorf("stop = %s", got)
		}
	})
	t.Run("not connected", func(t *testing.T) {
		if err := (&Device{}).StopApp(nil); !errors.Is(err, generic.ErrNotConnected) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("queue full", func(t *testing.T) {
		d := &Device{sendCh: make(chan []byte, 1)}
		_ = d.StopApp(nil)
		if err := d.StopApp(nil); !errors.Is(err, errSendQueueFull) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("unencodable command", func(t *testing.T) {
		if err := (&Device{}).sendCommand(make(chan int)); err == nil {
			t.Fatal("want an encode error")
		}
	})
}

func TestHandleManagerConnection(t *testing.T) {
	clock := testkit.NewFakeClock(epochTime())
	conn := newFakeConn()
	d := newDevice(infoOf("e1"), clock)
	d.conn = conn
	done := make(chan struct{})
	go func() { d.HandleManagerConnection(); close(done) }()

	eventually(t, "booted", func() bool { return d.DeviceState() == device.StateBooted })
	if err := d.StopApp(nil); err != nil {
		t.Fatal(err)
	}
	if got := string(<-conn.out); got != `{"Action":"stop"}` {
		t.Errorf("written = %s", got)
	}
	clock.Advance(time.Minute)
	conn.in <- []byte("ping")
	close(conn.in)
	<-done

	if d.DeviceState() != device.StateShutdown || !d.takeUpdated() {
		t.Error("a closed connection shuts the editor down and marks it updated")
	}
	if want := epochTime().Add(time.Minute); !d.lastUpdateAt.Equal(want) {
		t.Errorf("lastUpdateAt = %v, want %v", d.lastUpdateAt, want)
	}
	if err := d.StopApp(nil); !errors.Is(err, generic.ErrNotConnected) {
		t.Errorf("after close err = %v", err)
	}
}

func TestHandleManagerConnectionWriteError(t *testing.T) {
	conn := newFakeConn()
	conn.writeErr = errBoom
	d := newDevice(infoOf("e1"), testkit.NewFakeClock(epochTime()))
	d.conn = conn
	done := make(chan struct{})
	go func() { d.HandleManagerConnection(); close(done) }()

	eventually(t, "booted", func() bool { return d.DeviceState() == device.StateBooted })
	_ = d.StopApp(nil)
	<-conn.out
	close(conn.in)
	<-done
}

func TestHandleManagerConnectionWithoutConnection(t *testing.T) {
	d := newDevice(infoOf("e1"), nil)
	d.HandleManagerConnection()
	if d.DeviceState() != device.StateNone {
		t.Fatalf("state = %v, want unchanged", d.DeviceState())
	}
}

func TestUnityLogListening(t *testing.T) {
	d := newDevice(infoOf("e1"), testkit.NewFakeClock(epochTime()))
	d.UnityLogStopListening()
	d.UnityLogStartListening()
	if d.logCancel == nil {
		t.Fatal("listening has a cancel func")
	}
	d.UnityLogStopListening()
}

func TestProcessUnityLog(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "unity.log")
	if err := os.WriteFile(logFile, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticks := 0
	clock := newTickClock(t, func(time.Time) {
		if ticks++; ticks == 2 {
			cancel()
		}
	})
	d := newDevice(infoOf("e1"), clock)

	t.Run("missing file", func(t *testing.T) {
		d := newDevice(infoOf("e1"), newTickClock(t, nil))
		d.processUnityLog(context.Background(), filepath.Join(t.TempDir(), "none.log"))
	})
	t.Run("follows until cancelled", func(t *testing.T) {
		d.processUnityLog(ctx, logFile)
		if ticks < 2 {
			t.Fatalf("ticks = %d, want the start delay and one poll", ticks)
		}
	})
	t.Run("cancelled before the start delay", func(t *testing.T) {
		d := newDevice(infoOf("e1"), testkit.NewFakeClock(epochTime()))
		done, stop := context.WithCancel(context.Background())
		stop()
		d.processUnityLog(done, logFile)
	})
}

func TestFollowLogStopsOnReadError(t *testing.T) {
	d := newDevice(infoOf("e1"), testkit.NewFakeClock(epochTime()))
	d.followLog(context.Background(), bufio.NewReader(iotest.ErrReader(errBoom)))
}

func TestFollowLogIdleWaitsForTheClock(t *testing.T) {
	clock := testkit.NewFakeClock(epochTime())
	d := newDevice(infoOf("e1"), clock)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { d.followLog(ctx, bufio.NewReader(strings.NewReader("a\n"))); close(done) }()

	eventually(t, "idle wait", func() bool { return clock.Pending() == 1 })
	cancel()
	<-done
}
