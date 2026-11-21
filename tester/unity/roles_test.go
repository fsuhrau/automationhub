package unity

import (
	"context"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func nativeScriptResponse(script string) *appproto.Response {
	data := []byte(script)
	return &appproto.Response{Payload: appproto.ResponseData{Data: &data}}
}

func TestHandleNativeScriptRunsTheScriptOfADeviceThatHasNativeScripts(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	NewExecutor(nil, nil).handleNativeScript(dev, nativeScriptResponse("tap 1 2"))

	deadline := time.Now().Add(5 * time.Second) // the script runs in its own goroutine
	for len(dev.CallsTo("RunNativeScript")) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	calls := dev.CallsTo("RunNativeScript")
	if len(calls) != 1 || string(calls[0].Args[0].([]byte)) != "tap 1 2" {
		t.Fatalf("calls %v", calls)
	}
}

func TestHandleNativeScriptSkipsADeviceWithoutNativeScripts(t *testing.T) {
	hook := logtest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })

	dev := testkit.NewCoreDevice("d1")
	NewExecutor(nil, nil).handleNativeScript(dev, nativeScriptResponse("tap 1 2"))

	if len(dev.Fake.CallsTo("RunNativeScript")) != 0 {
		t.Fatal("the script must not run")
	}
	if entry := hook.LastEntry(); entry == nil || entry.Level != logrus.WarnLevel {
		t.Fatalf("want a warning, got %v", hook.Entries)
	}
}

func TestDeviceScreenshotNeedsTheScreenshotterRole(t *testing.T) {
	full := testkit.NewFakeDevice("d1")
	full.PNG = basetest.PNG()
	if got := deviceScreenshot(full); len(got) == 0 {
		t.Error("a device with a Screenshotter returns its screenshot")
	}
	if got := deviceScreenshot(testkit.NewCoreDevice("d2")); got != nil {
		t.Errorf("a core device has no native screenshot, got %d bytes", len(got))
	}
}

func TestHandleLogFinishesTheExecutionOnAnException(t *testing.T) {
	for name, tc := range map[string]struct {
		level    appproto.LogLevel
		finished bool
	}{
		"exception": {appproto.LogLevel_Exception, true},
		"error":     {appproto.LogLevel_Error, false},
	} {
		t.Run(name, func(t *testing.T) {
			e := NewExecutor(nil, nil)
			e.wg = sync.NewExtendedWaitGroup(context.Background())
			e.wg.Add(1)

			e.handleLog(testkit.NewFakeDevice("d1"), &appproto.Response{Payload: appproto.ResponseData{LogData: &appproto.LogData{Level: tc.level}}})

			finished := e.wg.WaitWithTimeout(20*time.Millisecond) == nil
			if finished != tc.finished {
				t.Fatalf("finished %v, want %v", finished, tc.finished)
			}
		})
	}
}
