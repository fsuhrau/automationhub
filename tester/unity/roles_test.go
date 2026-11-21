package unity

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func nativeScriptResponse(script string) *action.Response {
	data := []byte(script)
	return &action.Response{Payload: action.ResponseData{Data: &data}}
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
